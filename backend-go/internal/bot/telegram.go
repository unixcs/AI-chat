// Package bot implements the Telegram admin remote control. It talks to the
// service layer only (never SQL), allows a fixed admin-ID whitelist, and is
// entirely optional: without TG_BOT_TOKEN nothing starts.
//
// Mainland servers usually cannot reach api.telegram.org directly; set
// TG_PROXY (e.g. http://127.0.0.1:7890) to route the bot traffic through a
// local proxy. Connection failures are logged and retried — they never affect
// the chat service.
package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"ai-chat-backend/internal/config"
	"ai-chat-backend/internal/service"
)

type telegramUpdate struct {
	UpdateID int `json:"update_id"`
	Message  struct {
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
		From struct {
			ID       int64  `json:"id"`
			Username string `json:"username"`
		} `json:"from"`
		Text string `json:"text"`
	} `json:"message"`
}

type Bot struct {
	cfg   *config.Config
	svc   *service.Service
	hc    *http.Client
	offset int64
}

// Start launches the polling loop in a goroutine when configured.
func Start(ctx context.Context, cfg *config.Config, svc *service.Service) {
	if cfg.TGBotToken == "" {
		return
	}
	if len(cfg.TGAdminIDs) == 0 {
		log.Println("[tg-bot] TG_BOT_TOKEN 已配置但 TG_ADMIN_IDS 为空，bot 不启动")
		return
	}
	transport := &http.Transport{}
	if cfg.TGProxy != "" {
		proxyURL, err := url.Parse(cfg.TGProxy)
		if err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	} else {
		transport.Proxy = http.ProxyFromEnvironment
	}
	b := &Bot{
		cfg: cfg,
		svc: svc,
		hc: &http.Client{
			Timeout:   65 * time.Second,
			Transport: transport,
		},
	}
	go b.run(ctx)
	log.Printf("[tg-bot] started (admins=%d, proxy=%q)", len(cfg.TGAdminIDs), cfg.TGProxy)
}

func (b *Bot) run(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}
		updates, err := b.poll(ctx)
		if err != nil {
			log.Printf("[tg-bot] poll error: %v", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(15 * time.Second):
			}
			continue
		}
		for _, u := range updates {
			b.offset = int64(u.UpdateID) + 1
			b.handle(u)
		}
	}
}

func (b *Bot) poll(ctx context.Context) ([]telegramUpdate, error) {
	api := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?timeout=55&offset=%d", b.cfg.TGBotToken, b.offset)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, api, nil)
	if err != nil {
		return nil, err
	}
	resp, err := b.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var payload struct {
		OK     bool             `json:"ok"`
		Result []telegramUpdate `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if !payload.OK {
		return nil, fmt.Errorf("getUpdates not ok")
	}
	return payload.Result, nil
}

func (b *Bot) handle(u telegramUpdate) {
	chatID := u.Message.Chat.ID
	fromID := u.Message.From.ID
	if !b.isAdmin(fromID) {
		log.Printf("[tg-audit] operator=tg:%d action=UNAUTHORIZED cmd=%q", fromID, u.Message.Text)
		b.send(chatID, "未授权的 Telegram 账号。")
		return
	}
	text := strings.TrimSpace(u.Message.Text)
	if text == "" {
		return
	}
	log.Printf("[tg-audit] operator=tg:%d cmd=%q", fromID, text)
	b.send(chatID, b.dispatch(text))
}

func (b *Bot) isAdmin(id int64) bool {
	for _, allowed := range b.cfg.TGAdminIDs {
		if allowed == id {
			return true
		}
	}
	return false
}

func (b *Bot) dispatch(text string) string {
	parts := strings.Fields(text)
	cmd := strings.ToLower(strings.TrimPrefix(parts[0], "/"))
	arg := strings.TrimSpace(strings.TrimPrefix(text, parts[0]))

	switch cmd {
	case "start", "help":
		return "可用命令：\n/status 服务状态\n/ai AI 网关状态\n/code 生成 1 个邀请码\n/codes <n> 生成 n 个邀请码\n/user <手机号> 查询用户\n/ban <手机号> 封禁\n/unban <手机号> 解封\n/announce <文本> 发布一次性公告\n/reg on|off 开关注册\n/redeem on|off 开关兑换"
	case "status":
		return b.cmdStatus()
	case "ai":
		return b.cmdAI()
	case "code":
		return b.cmdCodes(1)
	case "codes":
		n := 1
		if v, err := strconv.Atoi(arg); err == nil && v > 0 {
			n = v
		}
		return b.cmdCodes(n)
	case "user":
		return b.cmdUser(arg)
	case "ban":
		return b.cmdSetStatus(arg, "disabled")
	case "unban":
		return b.cmdSetStatus(arg, "active")
	case "announce":
		return b.cmdAnnounce(arg)
	case "reg":
		return b.cmdToggle("注册", arg, b.svc.SetRegistrationOpen, b.svc.RegistrationOpen())
	case "redeem":
		return b.cmdToggle("兑换", arg, b.svc.SetRedeemOpen, b.svc.RedeemOpen())
	default:
		return "未知命令，发送 /help 查看列表。"
	}
}

func (b *Bot) cmdStatus() string {
	dash, err := b.svc.Dashboard()
	if err != nil {
		return "查询失败"
	}
	return fmt.Sprintf("服务正常 ✅\n用户数: %v\n有效会员: %v\n今日兑换: %v\n会话总数: %v\n注册: %v\n兑换: %v",
		dash["userCount"], dash["memberCount"], dash["todayRedeemCount"], dash["conversationCount"],
		boolText(b.svc.RegistrationOpen()), boolText(b.svc.RedeemOpen()))
}

func (b *Bot) cmdAI() string {
	models := b.svc.Router.HealthSnapshot()
	stats := b.svc.Router.StatsSnapshot()
	var sb strings.Builder
	fmt.Fprintf(&sb, "AI 模式: %s\n总请求: %v ｜ 官方兜底: %v ｜ 切换次数: %v\n", b.svc.Router.Mode(), stats["totalRequests"], stats["officialFallback"], stats["switches"])
	for _, m := range models {
		state := "✅"
		if !m.Available {
			state = "❄️冷却中"
		}
		fmt.Fprintf(&sb, "%s %s (%s) 成功%d/失败%d/超时%d\n", state, m.Name, m.Model, m.Successes, m.Failures, m.Timeouts)
	}
	return sb.String()
}

func (b *Bot) cmdCodes(n int) string {
	codes, serr := b.svc.GenerateRedeemCodes(n, 1)
	if serr != nil {
		return "生成失败：" + serr.Message
	}
	var sb strings.Builder
	sb.WriteString("已生成邀请码（30 天）：\n")
	for _, c := range codes {
		sb.WriteString(c.Code + "\n")
	}
	return sb.String()
}

func (b *Bot) cmdUser(phone string) string {
	user, err := b.svc.LookupUserByPhone(strings.TrimSpace(phone))
	if err != nil {
		return "用户不存在"
	}
	member := "未开通"
	if service.MembershipValid(user) {
		member = user.MemberExpireAt[:10] + " 到期"
	}
	return fmt.Sprintf("手机号: %s\n昵称: %s\n状态: %s\n会员: %s", user.Phone, user.Nickname, user.Status, member)
}

func (b *Bot) cmdSetStatus(phone, status string) string {
	user, err := b.svc.LookupUserByPhone(strings.TrimSpace(phone))
	if err != nil {
		return "用户不存在"
	}
	if serr := b.svc.AdminSetUserStatus(user.ID, status); serr != nil {
		return "操作失败：" + serr.Message
	}
	if status == "disabled" {
		return "已封禁 " + phone
	}
	return "已解封 " + phone
}

func (b *Bot) cmdAnnounce(text string) string {
	if strings.TrimSpace(text) == "" {
		return "用法: /announce <公告内容>"
	}
	_, serr := b.svc.AdminCreateAnnouncement("系统公告", text, true)
	if serr != nil {
		return "发布失败：" + serr.Message
	}
	return "公告已发布 ✅（用户确认后不再显示）"
}

func (b *Bot) cmdToggle(name, arg string, setter func(bool) error, current bool) string {
	switch strings.ToLower(strings.TrimSpace(arg)) {
	case "on", "open", "开":
		_ = setter(true)
		return name + "已开启"
	case "off", "close", "关":
		_ = setter(false)
		return name + "已关闭"
	default:
		return name + "当前: " + boolText(current) + "（用法 /" + strings.ToLower(name) + " on|off）"
	}
}

func boolText(b bool) string {
	if b {
		return "开启"
	}
	return "关闭"
}

func (b *Bot) send(chatID int64, text string) {
	if text == "" {
		return
	}
	payload, _ := json.Marshal(map[string]any{
		"chat_id": chatID,
		"text":    text,
	})
	api := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", b.cfg.TGBotToken)
	resp, err := b.hc.Post(api, "application/json", strings.NewReader(string(payload)))
	if err != nil {
		log.Printf("[tg-bot] send error: %v", err)
		return
	}
	resp.Body.Close()
}
