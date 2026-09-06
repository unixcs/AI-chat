package bot

import (
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"ai-chat-backend/internal/config"
	"ai-chat-backend/internal/service"
	"ai-chat-backend/internal/store"
)

func newBot(t *testing.T) *Bot {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "t.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { st.DB.Close() })
	cfg := &config.Config{SystemPrompt: "出厂默认", SystemPromptDefault: "出厂默认", RegistrationOpen: true, RedeemOpen: true}
	return &Bot{cfg: cfg, svc: service.New(cfg, st, nil), hc: &http.Client{}}
}

const testAdminID = 42

// T7: /prompt reflects the same service state the Web API sees, and
// /prompt set <v> switches to a copy of that version (new head).
func TestBotPromptStatusAndSwitch(t *testing.T) {
	b := newBot(t)

	out := b.dispatch(testAdminID, "prompt")
	if !strings.Contains(out, "未建立") || !strings.Contains(out, "env") {
		t.Fatalf("empty prompt status: %q", out)
	}

	if _, serr := b.svc.AdminUpdatePrompt("第一版", "tester"); serr != nil {
		t.Fatalf("update prompt: %+v", serr)
	}
	out = b.dispatch(testAdminID, "prompt")
	if !strings.Contains(out, "v1") || !strings.Contains(out, "db") {
		t.Fatalf("prompt status: %q", out)
	}

	if out := b.dispatch(testAdminID, "prompt set 1"); !strings.Contains(out, "v2") {
		t.Fatalf("switch output: %q", out)
	}
	info := b.svc.AdminPromptInfo()
	if info["version"] != int64(2) || info["content"] != "第一版" {
		t.Fatalf("after switch: %+v", info)
	}

	// T13: unknown version error copy
	if out := b.dispatch(testAdminID, "prompt set 99"); !strings.Contains(out, "切换失败") || !strings.Contains(out, "不存在") {
		t.Fatalf("unknown version: %q", out)
	}
	if out := b.dispatch(testAdminID, "prompt set abc"); !strings.Contains(out, "用法") {
		t.Fatalf("bad arg: %q", out)
	}
}

// T8: /codeinfo covers unused / used / void / missing; /redeemlog lists records.
func TestBotRedeemQueries(t *testing.T) {
	b := newBot(t)

	codes, serr := b.svc.GenerateRedeemCodes(2, 1, "tester")
	if serr != nil || len(codes) != 2 {
		t.Fatalf("generate: %+v", serr)
	}
	if out := b.dispatch(testAdminID, "codeinfo "+codes[0].Code); !strings.Contains(out, "未使用") {
		t.Fatalf("unused codeinfo: %q", out)
	}

	if serr := b.svc.Register("13911112222", "小明", "pass123"); serr != nil {
		t.Fatalf("register: %+v", serr)
	}
	user, err := b.svc.LookupUserByPhone("13911112222")
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if _, serr = b.svc.Redeem(user.ID, codes[0].Code); serr != nil {
		t.Fatalf("redeem: %+v", serr)
	}
	out := b.dispatch(testAdminID, "codeinfo "+codes[0].Code)
	if !strings.Contains(out, "已使用") || !strings.Contains(out, "13911112222") || !strings.Contains(out, "会员至") {
		t.Fatalf("used codeinfo: %q", out)
	}

	out = b.dispatch(testAdminID, "redeemlog")
	if !strings.Contains(out, codes[0].Code) || !strings.Contains(out, "13911112222") {
		t.Fatalf("redeemlog: %q", out)
	}

	if out := b.dispatch(testAdminID, "codeinfo VIP-NOPE"); !strings.Contains(out, "不存在") {
		t.Fatalf("missing code: %q", out)
	}
}

// T9: toggles echo current state on bad args without changing anything, and
// surface setter failures instead of swallowing them.
func TestBotToggleSemantics(t *testing.T) {
	b := newBot(t)

	before := b.svc.RegistrationOpen()
	out := b.dispatch(testAdminID, "reg")
	if !strings.Contains(out, "注册当前") || !strings.Contains(out, "/reg on|off") {
		t.Fatalf("state echo: %q", out)
	}
	if b.svc.RegistrationOpen() != before {
		t.Fatal("no-arg toggle must not change state")
	}
	out = b.dispatch(testAdminID, "reg xyz")
	if !strings.Contains(out, "注册当前") {
		t.Fatalf("bad-arg echo: %q", out)
	}

	if out := b.dispatch(testAdminID, "reg off"); !strings.Contains(out, "已关闭") {
		t.Fatalf("reg off: %q", out)
	}
	if b.svc.RegistrationOpen() {
		t.Fatal("reg off must close registration")
	}

	// setter failure surfaces: with the DB closed the write must error out
	if err := b.svc.Store.DB.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	if out := b.dispatch(testAdminID, "reg on"); !strings.Contains(out, "设置失败") {
		t.Fatalf("setter failure must surface: %q", out)
	}
}

// /user renders three membership states: none / valid / expired.
func TestBotUserMembershipStates(t *testing.T) {
	b := newBot(t)

	if serr := b.svc.Register("13911112222", "小明", "pass123"); serr != nil {
		t.Fatalf("register: %+v", serr)
	}
	user, err := b.svc.LookupUserByPhone("13911112222")
	if err != nil {
		t.Fatalf("lookup: %v", err)
	}
	if out := b.dispatch(testAdminID, "user 13911112222"); !strings.Contains(out, "未开通") {
		t.Fatalf("no membership: %q", out)
	}
	if _, serr := b.svc.AdminSetMemberExpire(user.ID, "2020-01-01T00:00:00Z", "tester"); serr != nil {
		t.Fatalf("set expire past: %+v", serr)
	}
	if out := b.dispatch(testAdminID, "user 13911112222"); !strings.Contains(out, "已过期") {
		t.Fatalf("expired: %q", out)
	}
	if _, serr := b.svc.AdminSetMemberExpire(user.ID, "2099-01-01T00:00:00Z", "tester"); serr != nil {
		t.Fatalf("set expire future: %+v", serr)
	}
	if out := b.dispatch(testAdminID, "user 13911112222"); !strings.Contains(out, "有效期至") {
		t.Fatalf("valid: %q", out)
	}
	if out := b.dispatch(testAdminID, "user 19900000000"); !strings.Contains(out, "用户不存在") {
		t.Fatalf("missing user: %q", out)
	}
}

// T12: overlong replies are truncated under the Telegram limit.
func TestBotMessageTruncation(t *testing.T) {
	long := strings.Repeat("字", maxTelegramMessage+500)
	got := truncateMessage(long)
	if len(got) > maxTelegramMessage+100 || !strings.Contains(got, "已截断") {
		t.Fatalf("truncation broken: %d chars", len(got))
	}
	if got := truncateMessage("短的"); got != "短的" {
		t.Fatalf("short text must pass through: %q", got)
	}
}

// R11 #2: the byte-budget cut must never split a multi-byte character — the
// truncated payload has to stay valid UTF-8 end to end.
func TestBotTruncationRuneSafe(t *testing.T) {
	cjk := strings.Repeat("测", 1334) // 4002 bytes, crosses the budget mid-rune
	got := truncateMessage(cjk)
	if !utf8.ValidString(got) {
		t.Fatalf("truncated output must be valid UTF-8")
	}
	if !utf8.ValidString(truncateMessage(strings.Repeat("\U0001F600", 1100))) {
		t.Fatalf("astral-plane truncation must stay valid UTF-8")
	}
	if got := truncateMessage(strings.Repeat("a", maxTelegramMessage+1)); !utf8.ValidString(got) || !strings.Contains(got, "已截断") {
		t.Fatalf("ascii truncation broken: %q tail", got[maxTelegramMessage-10:])
	}
}

// whitelist still enforced: unknown Telegram IDs are rejected.
func TestBotWhitelist(t *testing.T) {
	b := newBot(t)
	b.cfg.TGAdminIDs = []int64{42}
	if b.isAdmin(42) != true || b.isAdmin(43) != false {
		t.Fatal("whitelist broken")
	}
}
