package api

import (
	"strings"
	"testing"
)

// renameFixture registers and logs in one user, then creates a conversation.
// Conversational metadata operations (list/create/delete/rename) intentionally
// do not require membership, matching the legacy Node backend.
func renameFixture(t *testing.T, e *env, phone string) (token, conversationID string) {
	t.Helper()
	if code, payload := e.req("POST", "/api/auth/register", "", map[string]any{
		"phone": phone, "nickname": "rename", "password": "pass123",
	}); code != 200 {
		t.Fatalf("register %s: %d %v", phone, code, payload)
	}
	code, payload := e.req("POST", "/api/auth/login", "", map[string]any{"phone": phone, "password": "pass123"})
	if code != 200 {
		t.Fatalf("login %s: %d %v", phone, code, payload)
	}
	token = dataOf(payload)["token"].(string)
	if code, payload = e.req("POST", "/api/chat/conversations", token, nil); code != 200 {
		t.Fatalf("create conversation: %d %v", code, payload)
	}
	return token, dataOf(payload)["id"].(string)
}

func TestRenameConversation(t *testing.T) {
	env := newEnv(t, []string{"ok"})
	token, convID := renameFixture(t, env, "13922223333")

	// Success: the title is trimmed and echoed back with the conversation.
	code, payload := env.req("PUT", "/api/chat/conversations/"+convID+"/title", token, map[string]any{
		"title": "  星盘解读  ",
	})
	if code != 200 {
		t.Fatalf("rename: %d %v", code, payload)
	}
	data := dataOf(payload)
	if data["id"] != convID || data["title"] != "星盘解读" {
		t.Fatalf("rename response mismatch: %v", data)
	}
	if data["updatedAt"] == nil || data["updatedAt"] == "" {
		t.Fatalf("rename must bump updatedAt: %v", data)
	}

	// The list endpoint serves the new title.
	code, payload = env.req("GET", "/api/chat/conversations?page=1&pageSize=10", token, nil)
	if code != 200 {
		t.Fatalf("list: %d %v", code, payload)
	}
	found := false
	for _, item := range listOf(payload) {
		row := item.(map[string]any)
		if row["id"] == convID {
			found = true
			if row["title"] != "星盘解读" {
				t.Fatalf("list title not updated: %v", row)
			}
		}
	}
	if !found {
		t.Fatal("renamed conversation missing from list")
	}

	// Exactly 50 runes is accepted, 51 is rejected.
	code, _ = env.req("PUT", "/api/chat/conversations/"+convID+"/title", token, map[string]any{
		"title": strings.Repeat("星", 50),
	})
	if code != 200 {
		t.Fatalf("50-rune title must be accepted, got %d", code)
	}
	code, payload = env.req("PUT", "/api/chat/conversations/"+convID+"/title", token, map[string]any{
		"title": strings.Repeat("星", 51),
	})
	if code != 400 || payload["message"] != "标题过长，最多 50 个字符" {
		t.Fatalf("51-rune title must be 400: %d %v", code, payload)
	}

	// Blank/whitespace-only titles are rejected and leave the old title intact.
	for _, blank := range []string{"", "   "} {
		code, payload = env.req("PUT", "/api/chat/conversations/"+convID+"/title", token, map[string]any{"title": blank})
		if code != 400 || payload["message"] != "标题不能为空" {
			t.Fatalf("blank title %q must be 400: %d %v", blank, code, payload)
		}
	}
	code, payload = env.req("GET", "/api/chat/conversations?page=1&pageSize=10", token, nil)
	if code != 200 {
		t.Fatalf("list after rejected renames: %d %v", code, payload)
	}
	for _, item := range listOf(payload) {
		if row := item.(map[string]any); row["id"] == convID && row["title"] != strings.Repeat("星", 50) {
			t.Fatalf("rejected rename must not change the title: %v", row)
		}
	}

	// Missing conversation → 404; another user's conversation → 404 too.
	if code, _ = env.req("PUT", "/api/chat/conversations/does-not-exist/title", token, map[string]any{"title": "x"}); code != 404 {
		t.Fatalf("unknown conversation must 404, got %d", code)
	}
	otherToken, _ := renameFixture(t, env, "13922224444")
	if code, payload = env.req("PUT", "/api/chat/conversations/"+convID+"/title", otherToken, map[string]any{"title": "x"}); code != 404 {
		t.Fatalf("foreign conversation must 404, got %d %v", code, payload)
	}

	// Authentication is required.
	if code, _ = env.req("PUT", "/api/chat/conversations/"+convID+"/title", "", map[string]any{"title": "x"}); code != 401 {
		t.Fatalf("unauthenticated rename must 401, got %d", code)
	}
}
