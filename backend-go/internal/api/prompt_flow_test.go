package api

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

func adminLogin(t *testing.T, e *env) string {
	t.Helper()
	code, payload := e.req("POST", "/api/admin/auth/login", "", map[string]any{"username": "admin", "password": "admin123"})
	if code != 200 {
		t.Fatalf("admin login: %d %v", code, payload)
	}
	return dataOf(payload)["token"].(string)
}

// T2: with no promptRevisions row, the effective prompt falls back to the
// startup config and source=env.
func TestPromptFallsBackToEnvConfig(t *testing.T) {
	env := newEnv(t, []string{"ok"})
	adminToken := adminLogin(t, env)

	code, payload := env.req("GET", "/api/admin/prompt", adminToken, nil)
	if code != 200 {
		t.Fatalf("get prompt: %d %v", code, payload)
	}
	data := dataOf(payload)
	if data["source"] != "env" || data["version"] != nil || data["content"] != "你是测试助手。" {
		t.Fatalf("env fallback wrong: %v", data)
	}
}

// T1: after PUT, the NEXT chat request must send the new prompt to upstream.
func TestPromptUpdateAppliesToNextChat(t *testing.T) {
	upstream, cap := fakeOpenAICapture(t, []string{"星"})
	env := newEnvOn(t, upstream)
	adminToken := adminLogin(t, env)

	code, payload := env.req("PUT", "/api/admin/prompt", adminToken, map[string]any{"content": "新提示词ABC"})
	if code != 200 {
		t.Fatalf("put prompt: %d %v", code, payload)
	}
	data := dataOf(payload)
	if data["source"] != "db" || data["version"] != float64(1) || data["content"] != "新提示词ABC" {
		t.Fatalf("put result wrong: %v", data)
	}

	// mint code → register → redeem → chat
	code, payload = env.req("POST", "/api/admin/redeem-codes/batch", adminToken, map[string]any{"quantity": 1, "durationMonths": 1})
	if code != 200 {
		t.Fatalf("code batch: %d %v", code, payload)
	}
	inviteCode := listOf(payload)[0].(map[string]any)["code"].(string)
	if code, _ = env.req("POST", "/api/auth/register", "", map[string]any{"phone": "13911112222", "nickname": "小明", "password": "pass123"}); code != 200 {
		t.Fatal("register failed")
	}
	code, payload = env.req("POST", "/api/auth/login", "", map[string]any{"phone": "13911112222", "password": "pass123"})
	if code != 200 {
		t.Fatalf("login: %d %v", code, payload)
	}
	userToken := dataOf(payload)["token"].(string)
	code, _ = env.req("POST", "/api/user/redeem", userToken, map[string]any{"code": inviteCode})
	if code != 200 {
		t.Fatalf("redeem: %d", code)
	}
	code, payload = env.req("POST", "/api/chat/conversations", userToken, nil)
	if code != 200 {
		t.Fatalf("conversation: %d %v", code, payload)
	}
	convID := dataOf(payload)["id"].(string)
	st, events := env.sseStream(userToken, convID, "问题")
	if st != 200 || len(events) != 2 || events[0] != "星" || events[1] != "[DONE]" {
		t.Fatalf("stream: %d %v", st, events)
	}
	got := cap.systemMessage()
	if !strings.HasPrefix(got, "新提示词ABC") {
		t.Fatalf("upstream must receive the updated base prompt, got %q", got)
	}
	// 默认语义（Plan F5）：基础 Prompt 之后按序拼入三段偏好卡片
	for _, seg := range []string{"回答长度要求：", "回答风格要求：", "输出格式要求："} {
		if !strings.Contains(got, seg) {
			t.Fatalf("system prompt missing pref segment %q, got %q", seg, got)
		}
	}
}

// T3: a request that already started keeps the prompt it started with, even
// when the admin saves a new one while the stream is waiting for upstream.
func TestPromptFrozenForInFlightStream(t *testing.T) {
	upstream, cap := fakeOpenAICapture(t, []string{"星"})
	env := newEnvOn(t, upstream)
	adminToken := adminLogin(t, env)

	// set prompt v1 and prepare a membership user
	if code, _ := env.req("PUT", "/api/admin/prompt", adminToken, map[string]any{"content": "第一版提示词"}); code != 200 {
		t.Fatal("put v1 failed")
	}
	code, payload := env.req("POST", "/api/admin/redeem-codes/batch", adminToken, map[string]any{"quantity": 1, "durationMonths": 1})
	if code != 200 {
		t.Fatalf("code batch: %d", code)
	}
	inviteCode := listOf(payload)[0].(map[string]any)["code"].(string)
	env.req("POST", "/api/auth/register", "", map[string]any{"phone": "13911112222", "nickname": "小明", "password": "pass123"})
	code, payload = env.req("POST", "/api/auth/login", "", map[string]any{"phone": "13911112222", "password": "pass123"})
	userToken := dataOf(payload)["token"].(string)
	env.req("POST", "/api/user/redeem", userToken, map[string]any{"code": inviteCode})
	code, payload = env.req("POST", "/api/chat/conversations", userToken, nil)
	convID := dataOf(payload)["id"].(string)

	// hold the upstream's first byte so the stream stays in flight
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

	// admin saves v2 while the request is already in flight
	if code, _ := env.req("PUT", "/api/admin/prompt", adminToken, map[string]any{"content": "第二版提示词"}); code != 200 {
		t.Fatal("put v2 failed")
	}
	close(gate)
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("stream did not finish")
	}
	if got := cap.systemMessage(); !strings.HasPrefix(got, "第一版提示词") {
		t.Fatalf("in-flight request must keep its starting prompt, got %q", got)
	}
}

// T5: validation contract — empty/too long content, unknown restore version.
func TestPromptValidationContract(t *testing.T) {
	env := newEnv(t, []string{"ok"})
	adminToken := adminLogin(t, env)

	code, payload := env.req("PUT", "/api/admin/prompt", adminToken, map[string]any{"content": "   "})
	if code != 400 || payload["message"] != "提示词内容不能为空" {
		t.Fatalf("empty content: %d %v", code, payload)
	}
	code, payload = env.req("PUT", "/api/admin/prompt", adminToken, map[string]any{"content": strings.Repeat("字", 20001)})
	if code != 400 || !strings.Contains(payload["message"].(string), "最多") {
		t.Fatalf("oversize content: %d %v", code, payload)
	}
	code, payload = env.req("POST", "/api/admin/prompt/restore", adminToken, map[string]any{"version": 999})
	if code != 404 || payload["message"] != "该版本不存在" {
		t.Fatalf("unknown version: %d %v", code, payload)
	}

	// revisions list after a successful save is newest-first
	if code, _ = env.req("PUT", "/api/admin/prompt", adminToken, map[string]any{"content": "版本一"}); code != 200 {
		t.Fatal("put failed")
	}
	code, payload = env.req("GET", "/api/admin/prompt/revisions", adminToken, nil)
	if code != 200 || len(listOf(payload)) != 1 {
		t.Fatalf("revisions: %d %v", code, payload)
	}
	if listOf(payload)[0].(map[string]any)["content"] != "版本一" {
		t.Fatalf("revisions content: %v", listOf(payload)[0])
	}
}
