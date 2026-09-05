package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ai-chat-backend/internal/ai"
)

// Full user journey: register → admin mints code → redeem → chat (SSE) →
// history persists → preferences → announcements → conversation delete.
func TestFullUserJourney(t *testing.T) {
	e := newEnv(t, []string{"星", "位", "是", "吉"})
	env := e

	// health
	if code, _ := env.req("GET", "/api/health", "", nil); code != 200 {
		t.Fatal("health must be 200")
	}

	// admin login (seeded admin/admin123)
	code, payload := env.req("POST", "/api/admin/auth/login", "", map[string]any{"username": "admin", "password": "admin123"})
	if code != 200 || dataOf(payload)["token"] == nil {
		t.Fatalf("admin login failed: %d %v", code, payload)
	}
	adminToken := dataOf(payload)["token"].(string)

	// admin password change with weak password rejected
	if code, _ = env.req("PUT", "/api/admin/auth/password", adminToken, map[string]any{"newPassword": "123"}); code != 400 {
		t.Fatal("weak admin password must be rejected")
	}

	// generate invite code
	code, payload = env.req("POST", "/api/admin/redeem-codes/batch", adminToken, map[string]any{"quantity": 2, "durationMonths": 1})
	if code != 200 || len(listOf(payload)) != 2 {
		t.Fatalf("code batch failed: %d %v", code, payload)
	}
	inviteCode := listOf(payload)[0].(map[string]any)["code"].(string)

	// register + login
	if code, payload = env.req("POST", "/api/auth/register", "", map[string]any{"phone": "13911112222", "nickname": "小明", "password": "pass123"}); code != 200 {
		t.Fatalf("register failed: %d %v", code, payload)
	}
	if code, payload = env.req("POST", "/api/auth/register", "", map[string]any{"phone": "bad", "nickname": "x", "password": "pass123"}); code != 400 {
		t.Fatal("bad phone must be rejected")
	}
	code, payload = env.req("POST", "/api/auth/login", "", map[string]any{"phone": "13911112222", "password": "wrong"})
	if code != 400 || payload["message"] != "密码错误" {
		t.Fatalf("wrong password handling: %d %v", code, payload)
	}
	code, payload = env.req("POST", "/api/auth/login", "", map[string]any{"phone": "13911112222", "password": "pass123"})
	if code != 200 {
		t.Fatalf("login failed: %d %v", code, payload)
	}
	userToken := dataOf(payload)["token"].(string)

	// chat before membership → 403
	code, payload = env.req("POST", "/api/chat/conversations", userToken, nil)
	if code != 200 {
		t.Fatalf("conversation create must not require membership: %d %v", code, payload)
	}
	convID := dataOf(payload)["id"].(string)
	if st, ev := env.sseStream(userToken, convID, "我的运势如何"); st != 403 {
		t.Fatalf("stream without membership must be 403, got %d %v", st, ev)
	}

	// redeem
	code, payload = env.req("POST", "/api/user/redeem", userToken, map[string]any{"code": inviteCode})
	if code != 200 {
		t.Fatalf("redeem failed: %d %v", code, payload)
	}
	profile := dataOf(payload)["profile"].(map[string]any)
	if profile["memberExpireAt"] == nil {
		t.Fatal("membership must be set after redeem")
	}
	// same code again → rejected
	if code, _ = env.req("POST", "/api/user/redeem", userToken, map[string]any{"code": inviteCode}); code != 400 {
		t.Fatal("double redeem must fail")
	}

	// session-check + profile
	if code, payload = env.req("GET", "/api/user/profile", userToken, nil); code != 200 {
		t.Fatalf("profile: %d %v", code, payload)
	}

	// SSE stream end-to-end
	st, events := env.sseStream(userToken, convID, "我的运势如何")
	if st != 200 {
		t.Fatalf("stream status %d", st)
	}
	joined := strings.Join(events, "")
	if !strings.Contains(joined, "星位是吉") {
		t.Fatalf("stream content mismatch: %q", joined)
	}
	if events[len(events)-1] != "[DONE]" {
		t.Fatal("stream must end with [DONE]")
	}

	// messages persisted
	code, payload = env.req("GET", "/api/chat/conversations/"+convID+"/messages", userToken, nil)
	if code != 200 || len(listOf(payload)) != 2 {
		t.Fatalf("messages after stream: %d %v", code, payload)
	}
	last := listOf(payload)[1].(map[string]any)
	if last["content"] != "星位是吉" {
		t.Fatalf("assistant reply not persisted: %v", last)
	}

	// conversation list pagination contract
	code, payload = env.req("GET", "/api/chat/conversations?page=1&pageSize=10", userToken, nil)
	if code != 200 || len(listOf(payload)) != 1 {
		t.Fatalf("conversation list: %d %v", code, payload)
	}

	// preferences update + reflected in profile
	if code, _ = env.req("PUT", "/api/user/preferences", userToken, map[string]any{"answerLength": "concise", "answerStyle": "plain"}); code != 200 {
		t.Fatal("preferences update failed")
	}
	if code, payload = env.req("GET", "/api/user/profile", userToken, nil); code != 200 {
		t.Fatal("profile reload failed")
	}
	profile = dataOf(payload)
	if profile["answerLength"] != "concise" || profile["answerStyle"] != "plain" {
		t.Fatalf("preferences not persisted: %v", profile)
	}
	if code, _ = env.req("PUT", "/api/user/preferences", userToken, map[string]any{"answerLength": "bogus"}); code != 400 {
		t.Fatal("bogus preference must be rejected")
	}

	// announcement: admin publishes → user sees once → ack → gone
	if code, _ = env.req("POST", "/api/admin/announcements", adminToken, map[string]any{"title": "维护", "content": "今晚 02:00 维护", "active": true}); code != 200 {
		t.Fatal("announcement create failed")
	}
	code, payload = env.req("GET", "/api/announcements/current", userToken, nil)
	if code != 200 || dataOf(payload) == nil {
		t.Fatalf("announcement current: %d %v", code, payload)
	}
	annID := dataOf(payload)["id"].(string)
	if code, _ = env.req("POST", "/api/announcements/"+annID+"/ack", userToken, nil); code != 200 {
		t.Fatal("ack failed")
	}
	if code, payload = env.req("GET", "/api/announcements/current", userToken, nil); code != 200 || dataOf(payload) != nil {
		t.Fatalf("announcement must disappear after ack: %d %v", code, payload)
	}

	// delete conversation cascades
	if code, _ = env.req("DELETE", "/api/chat/conversations/"+convID, userToken, nil); code != 200 {
		t.Fatal("delete conversation failed")
	}
	if code, _ = env.req("GET", "/api/chat/conversations/"+convID+"/messages", userToken, nil); code != 404 {
		t.Fatal("deleted conversation must 404")
	}
}

func TestSessionKickAndAuthGuards(t *testing.T) {
	e := newEnv(t, []string{"ok"})
	env := e

	_, payload := env.req("POST", "/api/auth/register", "", map[string]any{"phone": "13911110000", "nickname": "A", "password": "pass123"})
	if _, payload = env.req("POST", "/api/auth/login", "", map[string]any{"phone": "13911110000", "password": "pass123"}); payload == nil {
		t.Fatal("login payload nil")
	}
	token1 := dataOf(payload)["token"].(string)

	// login again on another device → token1 kicked
	_, payload = env.req("POST", "/api/auth/login", "", map[string]any{"phone": "13911110000", "password": "pass123"})
	token2 := dataOf(payload)["token"].(string)

	code, payload := env.req("GET", "/api/user/profile", token1, nil)
	if code != 401 || payload["code"] != "SESSION_KICKED" {
		t.Fatalf("kicked session must 401 SESSION_KICKED, got %d %v", code, payload)
	}
	if code, _ = env.req("GET", "/api/user/profile", token2, nil); code != 200 {
		t.Fatal("fresh session must work")
	}
	// missing token
	if code, _ = env.req("GET", "/api/user/profile", "", nil); code != 401 {
		t.Fatal("missing token must 401")
	}
	// user token on admin endpoint
	if code, _ = env.req("GET", "/api/admin/dashboard", token2, nil); code != 403 {
		t.Fatal("user token on admin endpoint must 403")
	}
}

func TestAdminManagementEndpoints(t *testing.T) {
	e := newEnv(t, []string{"x"})
	env := e
	_, payload := env.req("POST", "/api/admin/auth/login", "", map[string]any{"username": "admin", "password": "admin123"})
	adminToken := dataOf(payload)["token"].(string)

	// register a user through the public endpoint
	env.req("POST", "/api/auth/register", "", map[string]any{"phone": "13900001111", "nickname": "B", "password": "pass123"})

	// users list with pagination
	code, payload := env.req("GET", "/api/admin/users?page=1&pageSize=10", adminToken, nil)
	if code != 200 {
		t.Fatalf("admin users: %d %v", code, payload)
	}
	items := dataOf(payload)["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("admin users items: %v", items)
	}
	userID := items[0].(map[string]any)["id"].(string)

	// status / reset password / member expire
	if code, _ = env.req("PUT", "/api/admin/users/"+userID+"/status", adminToken, map[string]any{"status": "disabled"}); code != 200 {
		t.Fatal("disable failed")
	}
	if code, _ = env.req("PUT", "/api/admin/users/"+userID+"/status", adminToken, map[string]any{"status": "active"}); code != 200 {
		t.Fatal("enable failed")
	}
	if code, _ = env.req("PUT", "/api/admin/users/"+userID+"/reset-password", adminToken, map[string]any{"newPassword": "newpass1"}); code != 200 {
		t.Fatal("reset failed")
	}
	if code, _ = env.req("PUT", "/api/admin/users/"+userID+"/member-expire-at", adminToken, map[string]any{"memberExpireAt": "2027-01-01 00:00"}); code != 200 {
		t.Fatal("member expire failed")
	}
	if code, _ = env.req("PUT", "/api/admin/users/"+userID+"/member-expire-at", adminToken, map[string]any{"memberExpireAt": "not-a-date"}); code != 400 {
		t.Fatal("bad date must be rejected")
	}

	// dashboard
	code, payload = env.req("GET", "/api/admin/dashboard", adminToken, nil)
	if code != 200 {
		t.Fatal("dashboard failed")
	}
	dash := dataOf(payload)
	if dash["userCount"].(float64) != 1 || dash["memberCount"].(float64) != 1 {
		t.Fatalf("dashboard counts wrong: %v", dash)
	}

	// redeem code void flow
	code, payload = env.req("POST", "/api/admin/redeem-codes/batch", adminToken, map[string]any{"quantity": 1})
	codeID := listOf(payload)[0].(map[string]any)["id"].(string)
	if code, _ = env.req("PUT", "/api/admin/redeem-codes/"+codeID+"/void", adminToken, nil); code != 200 {
		t.Fatal("void failed")
	}
	if code, _ = env.req("PUT", "/api/admin/redeem-codes/"+codeID+"/void", adminToken, nil); code != 400 {
		t.Fatal("second void must fail")
	}

	// settings toggles
	if code, _ = env.req("PUT", "/api/admin/settings", adminToken, map[string]any{"registrationOpen": false}); code != 200 {
		t.Fatal("settings update failed")
	}
	if code, payload = env.req("POST", "/api/auth/register", "", map[string]any{"phone": "13900002222", "nickname": "C", "password": "pass123"}); code != 403 {
		t.Fatalf("registration must be closed: %d %v", code, payload)
	}
	if code, _ = env.req("PUT", "/api/admin/settings", adminToken, map[string]any{"registrationOpen": true}); code != 200 {
		t.Fatal("settings restore failed")
	}

	// AI status
	code, payload = env.req("GET", "/api/admin/ai/status", adminToken, nil)
	if code != 200 || dataOf(payload)["mode"] != "official" {
		t.Fatalf("ai status: %d %v", code, payload)
	}
}

// Stop generation: client disconnects mid-stream → partial content persisted,
// no [DONE], upstream cancelled.
func TestStopGenerationPersistsPartial(t *testing.T) {
	e := newEnv(t, nil) // custom upstream below
	env := e
	gotFirst := make(chan struct{})
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher := w.(http.Flusher)
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, `data: {"choices":[{"delta":{"content":"PARTIAL"}}]}`+"\n\n")
		flusher.Flush()
		close(gotFirst)
		<-r.Context().Done() // hang until cancelled
	}))
	defer upstream.Close()
	env.cfg.AIProviders[0].BaseURL = upstream.URL
	router := ai.NewRouter(env.cfg)
	env.server.Config.Handler = New(env.svc, router).Routes()

	_, payload := env.req("POST", "/api/auth/register", "", map[string]any{"phone": "13933334444", "nickname": "Stop", "password": "pass123"})
	_, payload = env.req("POST", "/api/auth/login", "", map[string]any{"phone": "13933334444", "password": "pass123"})
	token := dataOf(payload)["token"].(string)
	// membership via admin
	_, ap := env.req("POST", "/api/admin/auth/login", "", map[string]any{"username": "admin", "password": "admin123"})
	admin := dataOf(ap)["token"].(string)
	_, cp := env.req("POST", "/api/admin/redeem-codes/batch", admin, map[string]any{"quantity": 1})
	code := listOf(cp)[0].(map[string]any)["code"].(string)
	env.req("POST", "/api/user/redeem", token, map[string]any{"code": code})
	_, conv := env.req("POST", "/api/chat/conversations", token, nil)
	convID := dataOf(conv)["id"].(string)

	ctx, cancel := context.WithCancel(context.Background())
	req, _ := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s/api/chat/conversations/%s/stream?token=%s&content=hi", env.server.URL, convID, token), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("stream: %v", err)
	}
	buf := make([]byte, 4096)
	var received string
	for {
		n, err := resp.Body.Read(buf)
		received += string(buf[:n])
		if err != nil || strings.Contains(received, "PARTIAL") {
			break
		}
	}
	select {
	case <-gotFirst:
	case <-time.After(2 * time.Second):
		t.Fatal("upstream never produced first token")
	}
	cancel() // 停止生成
	resp.Body.Close()
	<-time.After(300 * time.Millisecond)

	_, payload = env.req("GET", "/api/chat/conversations/"+convID+"/messages", token, nil)
	msgs := listOf(payload)
	if len(msgs) != 2 {
		t.Fatalf("partial reply must be persisted, got %d messages", len(msgs))
	}
	if msgs[1].(map[string]any)["content"] != "PARTIAL" {
		t.Fatalf("persisted partial content mismatch: %v", msgs[1])
	}
	if strings.Contains(received, "[DONE]") {
		t.Fatal("cancelled stream must not deliver [DONE]")
	}
}

// An admin token pointed at the SSE endpoint must get the Node-compatible
// 403 INVALID_USER_TOKEN — not a panic and not a 401.
func TestStreamRejectsAdminToken(t *testing.T) {
	e := newEnv(t, []string{"x"})
	env := e
	_, payload := env.req("POST", "/api/admin/auth/login", "", map[string]any{"username": "admin", "password": "admin123"})
	admin := dataOf(payload)["token"].(string)

	resp, err := http.Get(env.server.URL + "/api/chat/conversations/whatever/stream?token=" + admin + "&content=hi")
	if err != nil {
		t.Fatalf("stream: %v", err)
	}
	defer resp.Body.Close()
	var body map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&body)
	if resp.StatusCode != 403 || body["code"] != "INVALID_USER_TOKEN" {
		t.Fatalf("expected 403 INVALID_USER_TOKEN, got %d %v", resp.StatusCode, body)
	}
}
