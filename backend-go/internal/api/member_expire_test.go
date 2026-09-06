package api

import "testing"

// The admin UI's datetime-local input submits "T"-separated, zone-less
// wall-clock (+8). The endpoint must accept it, normalize it to UTC storage,
// and surface errors instead of leaving the admin dialog dead.
func TestAdminMemberExpireAcceptsDateTimeLocal(t *testing.T) {
	env := newEnv(t, []string{"ok"})

	code, payload := env.req("POST", "/api/admin/auth/login", "", map[string]any{"username": "admin", "password": "admin123"})
	if code != 200 {
		t.Fatalf("admin login: %d %v", code, payload)
	}
	adminToken := dataOf(payload)["token"].(string)

	if code, _ = env.req("POST", "/api/auth/register", "", map[string]any{"phone": "13911112222", "nickname": "小明", "password": "pass123"}); code != 200 {
		t.Fatal("register failed")
	}
	code, payload = env.req("GET", "/api/admin/users?page=1&pageSize=10&phone=13911112222", adminToken, nil)
	if code != 200 {
		t.Fatalf("user query: %d %v", code, payload)
	}
	items, ok := dataOf(payload)["items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("user query items: %v", dataOf(payload)["items"])
	}
	uid := items[0].(map[string]any)["id"].(string)
	path := "/api/admin/users/" + uid + "/member-expire-at"

	// datetime-local with seconds (step=1)
	code, payload = env.req("PUT", path, adminToken, map[string]any{"memberExpireAt": "2027-01-01T00:00:00"})
	if code != 200 {
		t.Fatalf("datetime-local value must be accepted: %d %v", code, payload)
	}
	if got := dataOf(payload)["memberExpireAt"]; got != "2026-12-31T16:00:00.000Z" {
		t.Fatalf("+8 normalization wrong: %v", got)
	}

	// datetime-local minutes-only
	code, payload = env.req("PUT", path, adminToken, map[string]any{"memberExpireAt": "2027-01-01T00:30"})
	if code != 200 || dataOf(payload)["memberExpireAt"] != "2026-12-31T16:30:00.000Z" {
		t.Fatalf("minutes-only value: %d %v", code, payload)
	}

	// garbage must 400 with the contract message
	code, payload = env.req("PUT", path, adminToken, map[string]any{"memberExpireAt": "not-a-date"})
	if code != 400 || payload["message"] != "会员到期时间格式不正确" {
		t.Fatalf("garbage value: %d %v", code, payload)
	}

	// null clears
	code, payload = env.req("PUT", path, adminToken, map[string]any{"memberExpireAt": nil})
	if code != 200 || dataOf(payload)["memberExpireAt"] != nil {
		t.Fatalf("null must clear: %d %v", code, payload)
	}
}
