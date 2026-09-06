package api

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

// registerRedeemedUser registers a user, redeems a fresh invite code and
// returns the user token.
func registerRedeemedUser(t *testing.T, e *env, adminToken, phone string) string {
	t.Helper()
	code, payload := e.req("POST", "/api/admin/redeem-codes/batch", adminToken, map[string]any{"quantity": 1, "durationMonths": 1})
	if code != 200 {
		t.Fatalf("code batch: %d %v", code, payload)
	}
	inviteCode := listOf(payload)[0].(map[string]any)["code"].(string)
	e.req("POST", "/api/auth/register", "", map[string]any{"phone": phone, "nickname": "偏好用户", "password": "pass123"})
	code, payload = e.req("POST", "/api/auth/login", "", map[string]any{"phone": phone, "password": "pass123"})
	if code != 200 {
		t.Fatalf("login: %d %v", code, payload)
	}
	userToken := dataOf(payload)["token"].(string)
	if code, _ := e.req("POST", "/api/user/redeem", userToken, map[string]any{"code": inviteCode}); code != 200 {
		t.Fatalf("redeem: %d", code)
	}
	return userToken
}

// P1/P2/P9 via HTTP: answerFormat round-trips, invalid values 400, partial
// PUT leaves other dimensions untouched.
func TestPreferencesAnswerFormat(t *testing.T) {
	env := newEnv(t, []string{"ok"})
	adminToken := adminLogin(t, env)
	userToken := registerRedeemedUser(t, env, adminToken, "13911112222")

	code, payload := env.req("PUT", "/api/user/preferences", userToken, map[string]any{
		"answerLength": "concise", "answerStyle": "plain", "answerFormat": "plain"})
	if code != 200 || payload["data"] != true {
		t.Fatalf("preferences PUT: %d %v", code, payload)
	}
	code, payload = env.req("GET", "/api/user/profile", userToken, nil)
	profile := dataOf(payload)
	if profile["answerFormat"] != "plain" || profile["answerLength"] != "concise" || profile["answerStyle"] != "plain" {
		t.Fatalf("profile must carry all three dimensions: %v", profile)
	}

	// partial PUT (only format) must not clear the other two
	code, _ = env.req("PUT", "/api/user/preferences", userToken, map[string]any{"answerFormat": "standard"})
	if code != 200 {
		t.Fatalf("partial PUT: %d", code)
	}
	_, payload = env.req("GET", "/api/user/profile", userToken, nil)
	profile = dataOf(payload)
	if profile["answerFormat"] != "standard" || profile["answerLength"] != "concise" || profile["answerStyle"] != "plain" {
		t.Fatalf("partial PUT must preserve other dimensions: %v", profile)
	}

	// invalid value → 400
	code, payload = env.req("PUT", "/api/user/preferences", userToken, map[string]any{"answerFormat": "markdown"})
	if code != 400 || payload["message"] != "输出格式参数错误" {
		t.Fatalf("invalid format: %d %v", code, payload)
	}
}

// P6: card endpoints are adminAuth'd, whitelist-validated and auditable.
func TestAdminPrefPromptsEndpoints(t *testing.T) {
	env := newEnv(t, []string{"ok"})
	adminToken := adminLogin(t, env)

	// no token → 401
	if code, _ := env.req("GET", "/api/admin/pref-prompts", "", nil); code != 401 {
		t.Fatalf("unauthenticated list must 401, got %d", code)
	}

	code, payload := env.req("GET", "/api/admin/pref-prompts", adminToken, nil)
	if code != 200 {
		t.Fatalf("list: %d %v", code, payload)
	}
	cards := listOf(payload)
	if len(cards) != 10 {
		t.Fatalf("10 cards expected, got %d", len(cards))
	}
	first := cards[0].(map[string]any)
	if first["dimension"] != "answerLength" || first["value"] != "concise" || first["label"] != "精简" {
		t.Fatalf("card shape wrong: %v", first)
	}
	if first["customized"] != false || !strings.Contains(first["content"].(string), "精简") {
		t.Fatalf("fresh card must be preset: %v", first)
	}

	// update
	code, _ = env.req("PUT", "/api/admin/pref-prompts/answerLength/concise", adminToken, map[string]any{"content": "五十个字以内说完"})
	if code != 200 {
		t.Fatalf("update: %d", code)
	}
	_, payload = env.req("GET", "/api/admin/pref-prompts", adminToken, nil)
	for _, c := range listOf(payload) {
		cm := c.(map[string]any)
		if cm["value"] == "concise" {
			if cm["content"] != "五十个字以内说完" || cm["customized"] != true {
				t.Fatalf("edited card wrong: %v", cm)
			}
		}
	}

	// reset → back to preset
	if code, _ = env.req("POST", "/api/admin/pref-prompts/answerLength/concise/reset", adminToken, nil); code != 200 {
		t.Fatalf("reset: %d", code)
	}
	_, payload = env.req("GET", "/api/admin/pref-prompts", adminToken, nil)
	for _, c := range listOf(payload) {
		cm := c.(map[string]any)
		if cm["value"] == "concise" && (cm["customized"] != false || !strings.Contains(cm["content"].(string), "精简")) {
			t.Fatalf("reset must restore preset: %v", cm)
		}
	}

	// whitelist: unknown dimension/value → 404
	if code, _ = env.req("PUT", "/api/admin/pref-prompts/answerTone/warm", adminToken, map[string]any{"content": "x"}); code != 404 {
		t.Fatalf("unknown dimension must 404, got %d", code)
	}
	if code, _ = env.req("POST", "/api/admin/pref-prompts/answerLength/tiny/reset", adminToken, nil); code != 404 {
		t.Fatalf("unknown value reset must 404, got %d", code)
	}
}

// P5: a pref-card edit made while a stream is in flight must not leak into
// that stream — the system message was frozen at request start.
func TestPrefCardFrozenForInFlightStream(t *testing.T) {
	upstream, cap := fakeOpenAICapture(t, []string{"星"})
	env := newEnvOn(t, upstream)
	adminToken := adminLogin(t, env)
	userToken := registerRedeemedUser(t, env, adminToken, "13911112222")

	code, payload := env.req("POST", "/api/chat/conversations", userToken, nil)
	if code != 200 {
		t.Fatalf("conversation: %d %v", code, payload)
	}
	convID := dataOf(payload)["id"].(string)

	// the user prefers 大白话 so the plain style card is in effect
	if code, _ := env.req("PUT", "/api/user/preferences", userToken, map[string]any{
		"answerLength": "standard", "answerStyle": "plain", "answerFormat": "standard"}); code != 200 {
		t.Fatal("set preferences failed")
	}

	cap.mu.Lock()
	cap.hold = make(chan struct{})
	gate := cap.hold
	cap.mu.Unlock()

	events := make(chan string, 8)
	done := make(chan struct{})
	go func() {
		defer close(done)
		url := fmt.Sprintf("%s/api/chat/conversations/%s/stream?token=%s&content=%s",
			env.server.URL, convID, userToken, "问题")
		resp, err := http.Get(url)
		if err != nil {
			events <- "STREAM_FAIL"
			return
		}
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "data:") {
				events <- strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			}
		}
	}()

	if !cap.waitBody(2 * time.Second) {
		t.Fatal("upstream never received the stream request")
	}
	before := cap.systemMessage()

	// edit the style card while the stream is held at the upstream gate
	if code, _ := env.req("PUT", "/api/admin/pref-prompts/answerStyle/plain", adminToken, map[string]any{"content": "半途改的文案不应生效"}); code != 200 {
		t.Fatal("mid-flight card edit failed")
	}
	close(gate)
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("stream did not finish")
	}
	after := cap.systemMessage()
	if after != before {
		t.Fatalf("in-flight stream must keep its starting cards:\n before=%q\n after=%q", before, after)
	}
	if strings.Contains(after, "半途改的文案") {
		t.Fatalf("mid-flight edit leaked into frozen stream: %q", after)
	}

	// the NEXT request uses the new card content
	code, payload = env.req("POST", "/api/chat/conversations", userToken, nil)
	if code != 200 {
		t.Fatalf("second conversation: %d %v", code, payload)
	}
	convID2 := dataOf(payload)["id"].(string)
	st, ev := env.sseStream(userToken, convID2, "问题")
	if st != 200 || len(ev) == 0 {
		t.Fatalf("second stream: %d %v", st, ev)
	}
	if got := cap.systemMessage(); !strings.Contains(got, "半途改的文案") {
		t.Fatalf("next request must use the edited card, got %q", got)
	}
}
