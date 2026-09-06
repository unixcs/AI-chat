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
	"unicode/utf8"

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
	b.send(chatID, b.dispatch(fromID, text))
}

func (b *Bot) isAdmin(id int64) bool {
	for _, allowed := range b.cfg.TGAdminIDs {
		if allowed == id {
			return true
		}
	}
	return false
}

func (b *Bot) dispatch(fromID int64, text string) string {
	parts := strings.Fields(text)
	cmd := strings.ToLower(strings.TrimPrefix(parts[0], "/"))
	arg := strings.TrimSpace(strings.TrimPrefix(text, parts[0]))

	switch cmd {
	case "start", "help":
		return "可用命令：\n/status 服务状态\n/ai AI 网关状态\n/code 生成 1 个邀请码（1 个月）\n/codes <n> 生成 n 个邀请码\n/user <手机号> 查询用户\n/ban <手机号> 封禁\n/unban <手机号> 解封\n/redeemlog [n] 最近 n 条兑换记录\n/codeinfo <邀请码> 查询单码状态\n/announce <文本> 发布一次性公告\n/reg on|off 开关注册\n/redeem on|off 开关兑换\n/prompt 查看内置提示词\n/prompt set <版本号> 切换提示词版本"
	case "status":
		return b.cmdStatus()
	case "ai":
		return b.cmdAI()
	case "code":
		return b.cmdCodes(fromID, 1)
	case "codes":
		n := 1
		if v, err := strconv.Atoi(arg); err == nil && v > 0 {
			n = v
		}
		return b.cmdCodes(fromID, n)
	case "user":
		return b.cmdUser(fromID, arg)
	case "ban":
		return b.cmdSetStatus(fromID, arg, "disabled")
	case "unban":
		return b.cmdSetStatus(fromID, arg, "active")
	case "redeemlog":
		return b.cmdRedeemLog(arg)
	case "codeinfo":
		return b.cmdCodeInfo(arg)
	case "prompt":
		return b.cmdPrompt(fromID, arg)
	case "announce":
		return b.cmdAnnounce(fromID, arg)
	case "reg":
		return b.cmdToggle("reg", "注册", arg, b.svc.RegistrationOpen(), func(open bool) error { return b.svc.SetRegistrationOpen(open, b.operator(fromID)) })
	case "redeem":
		return b.cmdToggle("redeem", "兑换", arg, b.svc.RedeemOpen(), func(open bool) error { return b.svc.SetRedeemOpen(open, b.operator(fromID)) })
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

func (b *Bot) cmdCodes(fromID int64, n int) string {
	codes, serr := b.svc.GenerateRedeemCodes(n, 1, b.operator(fromID))
	if serr != nil {
		return "生成失败：" + serr.Message
	}
	var sb strings.Builder
	sb.WriteString("已生成邀请码（1 个月，约 30 天）：\n")
	for _, c := range codes {
		sb.WriteString(c.Code + "\n")
	}
	return sb.String()
}

func (b *Bot) cmdUser(fromID int64, phone string) string {
	user, err := b.svc.LookupUserByPhone(strings.TrimSpace(phone))
	if err != nil {
		return "用户不存在"
	}
	// 三态：未开通 / 有效 / 已过期（MembershipValid 对后两者均返回 false，
	// 不能只靠它区分）
	member := "未开通"
	if user.MemberExpireAt != "" {
		if service.MembershipValid(user) {
			member = "有效期至 " + shortDate(user.MemberExpireAt)
		} else {
			member = "已过期（" + shortDate(user.MemberExpireAt) + "）"
		}
	}
	return fmt.Sprintf("手机号: %s\n昵称: %s\n状态: %s\n会员: %s", user.Phone, user.Nickname, user.Status, member)
}

// shortDate renders the date part of an ISO stamp defensively.
func shortDate(iso string) string {
	if len(iso) >= 10 {
		return iso[:10]
	}
	return iso
}

func (b *Bot) cmdSetStatus(fromID int64, phone, status string) string {
	user, err := b.svc.LookupUserByPhone(strings.TrimSpace(phone))
	if err != nil {
		return "用户不存在"
	}
	if serr := b.svc.AdminSetUserStatus(user.ID, status, b.operator(fromID)); serr != nil {
		return "操作失败：" + serr.Message
	}
	if status == "disabled" {
		return "已封禁 " + phone
	}
	return "已解封 " + phone
}

func (b *Bot) cmdAnnounce(fromID int64, text string) string {
	if strings.TrimSpace(text) == "" {
		return "用法: /announce <公告内容>"
	}
	_, serr := b.svc.AdminCreateAnnouncement("系统公告", text, true, b.operator(fromID))
	if serr != nil {
		return "发布失败：" + serr.Message
	}
	return "公告已发布 ✅（用户确认后不再显示）"
}

func (b *Bot) cmdRedeemLog(arg string) string {
	n := 5
	if v, err := strconv.Atoi(strings.TrimSpace(arg)); err == nil && v > 0 {
		n = v
	}
	records, serr := b.svc.BotRedeemRecords(n)
	if serr != nil {
		return "查询失败：" + serr.Message
	}
	if len(records) == 0 {
		return "还没有兑换记录。"
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "最近 %d 条兑换记录：\n", len(records))
	for _, r := range records {
		fmt.Fprintf(&sb, "%s → %s @ %s（会员至 %s）\n", r.Code, r.Phone, shortDate(r.ActivatedAt), shortDate(r.AfterExpireAt))
	}
	return sb.String()
}

func (b *Bot) cmdCodeInfo(arg string) string {
	info, serr := b.svc.BotRedeemCodeInfo(arg)
	if serr != nil {
		return serr.Message
	}
	return info
}

func (b *Bot) cmdPrompt(fromID int64, arg string) string {
	if strings.TrimSpace(arg) == "" {
		info := b.svc.AdminPromptInfo()
		version := "未建立（使用出厂/环境配置）"
		if v, ok := info["version"].(int64); ok {
			version = fmt.Sprintf("v%d", v)
		}
		updated := "-"
		if info["updatedAt"] != nil {
			updated = shortDate(fmt.Sprintf("%v", info["updatedAt"]))
		}
		return fmt.Sprintf("内置提示词\n版本: %s\n来源: %v\n最后修改: %s\n操作人: %v\n（/prompt set <版本号> 切换历史版本）",
			version, info["source"], updated, info["operator"])
	}
	fields := strings.Fields(arg)
	if len(fields) == 2 && strings.EqualFold(fields[0], "set") {
		v, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			return "用法: /prompt set <版本号>"
		}
		info, serr := b.svc.AdminRestorePrompt(v, b.operator(fromID))
		if serr != nil {
			return "切换失败：" + serr.Message
		}
		return fmt.Sprintf("已切换到版本 v%d ✅（新聊天立即生效）", info["version"])
	}
	return "用法: /prompt 或 /prompt set <版本号>"
}

func (b *Bot) cmdToggle(cmdName, label, arg string, current bool, setter func(bool) error) string {
	switch strings.ToLower(strings.TrimSpace(arg)) {
	case "on", "open", "开":
		if err := setter(true); err != nil {
			return label + "设置失败：" + err.Error()
		}
		return label + "已开启"
	case "off", "close", "关":
		if err := setter(false); err != nil {
			return label + "设置失败：" + err.Error()
		}
		return label + "已关闭"
	default:
		state := "关闭"
		if current {
			state = "开启"
		}
		return label + "当前: " + state + "（用法 /" + cmdName + " on|off）"
	}
}

// operator renders the audit identity for a Telegram admin command.
func (b *Bot) operator(fromID int64) string {
	return fmt.Sprintf("tg:%d", fromID)
}

func boolText(b bool) string {
	if b {
		return "开启"
	}
	return "关闭"
}

// maxTelegramMessage keeps replies under the sendMessage 4096-char limit.
// The budget counts bytes: any ≤4000-byte string is also ≤4000 UTF-16 units,
// no matter how many astral-plane characters it contains.
const maxTelegramMessage = 4000

// truncateMessage defensively cuts replies that exceed the limit. The cut
// point backs off to the nearest rune boundary so a multi-byte character is
// never split into invalid UTF-8 (which Telegram would render as U+FFFD).
func truncateMessage(text string) string {
	if len(text) <= maxTelegramMessage {
		return text
	}
	cut := maxTelegramMessage
	for cut > 0 && !utf8.RuneStart(text[cut]) {
		cut--
	}
	return text[:cut] + "\n…（内容过长已截断）"
}

func (b *Bot) send(chatID int64, text string) {
	if text == "" {
		return
	}
	text = truncateMessage(text)
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
