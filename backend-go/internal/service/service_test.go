package service

import (
	"path/filepath"
	"strings"
	"testing"

	"ai-chat-backend/internal/config"
	"ai-chat-backend/internal/model"
	"ai-chat-backend/internal/store"
)

func TestBuildContextByRounds(t *testing.T) {
	history := []model.Message{
		{Role: "user", Content: "q1", CreatedAt: "2026-01-01T00:00:01.000Z"},
		{Role: "assistant", Content: "a1", CreatedAt: "2026-01-01T00:00:02.000Z"},
		{Role: "user", Content: "q2", CreatedAt: "2026-01-01T00:00:03.000Z"},
		{Role: "assistant", Content: "a2", CreatedAt: "2026-01-01T00:00:04.000Z"},
		{Role: "user", Content: "q3", CreatedAt: "2026-01-01T00:00:05.000Z"},
	}
	ctx := buildContextByRounds(history, 2)
	// last 2 rounds = [q2,a2] + dangling [q3]
	joined := ""
	for _, m := range ctx {
		joined += m.Role + ":" + m.Content + " "
	}
	if joined != "user:q2 assistant:a2 user:q3 " {
		t.Fatalf("context rounds wrong: %q", joined)
	}
}

func TestAnswerModeSuffix(t *testing.T) {
	if got := answerModeSuffix("", ""); got != "" {
		t.Fatalf("empty prefs must yield empty suffix, got %q", got)
	}
	got := answerModeSuffix("concise", "plain")
	if !strings.Contains(got, "精简") || !strings.Contains(got, "大白话") {
		t.Fatalf("suffix missing keywords: %q", got)
	}
}

func TestNormalizeMemberExpireAt(t *testing.T) {
	// Naive values are admin wall-clock (+8): the admin UI's datetime-local
	// submits browser-local (+8) time and redisplays stored values the same
	// way, so +8 keeps the pick→display round trip WYSIWYG. Zoned values
	// honor their own offset.
	cases := map[string]string{
		"2027-01-01T00:00:00":        "2026-12-31T16:00:00.000Z",
		"2027-01-01T00:00":           "2026-12-31T16:00:00.000Z",
		"2027-01-01 08:30:00":        "2027-01-01T00:30:00.000Z",
		"2027-01-01":                 "2026-12-31T16:00:00.000Z",
		"2027-01-01T00:00:00Z":       "2027-01-01T00:00:00.000Z",
		"2027-01-01T08:00:00+08:00":  "2027-01-01T00:00:00.000Z",
		"":                           "",
	}
	for in, want := range cases {
		got, err := NormalizeMemberExpireAt(in)
		if err != nil {
			t.Fatalf("%q: %v", in, err)
		}
		if got != want {
			t.Fatalf("%q → %q, want %q", in, got, want)
		}
	}
	if _, err := NormalizeMemberExpireAt("garbage"); err == nil {
		t.Fatal("garbage must error")
	}
}

func TestAdminUpdateAnnouncementResetSemantics(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "t.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	svc := New(&config.Config{}, st, nil)
	if _, serr := svc.AdminCreateAnnouncement("维护通知", "今晚维护", true, "test"); serr != nil {
		t.Fatalf("create: %+v", serr)
	}
	list, serr := svc.AdminListAnnouncements()
	if serr != nil || len(list) != 1 {
		t.Fatalf("list: %+v", serr)
	}
	id := list[0]["id"].(string)
	if err := st.AckAnnouncement(id, "u1", ""); err != nil {
		t.Fatalf("ack: %v", err)
	}

	// active-only toggle: users who already read it are NOT re-pinged
	if serr := svc.AdminUpdateAnnouncement(id, "", "", boolPtr(false), "test"); serr != nil {
		t.Fatalf("active toggle: %+v", serr)
	}
	if _, err := st.CurrentUnreadAnnouncement("u1"); err != store.ErrNotFound {
		t.Fatalf("active-only toggle must keep read state, got %v", err)
	}

	// content change (re-activating): every user gets re-notified
	if serr := svc.AdminUpdateAnnouncement(id, "", "改期到明晚维护", boolPtr(true), "test"); serr != nil {
		t.Fatalf("content update: %+v", serr)
	}
	got, err := st.CurrentUnreadAnnouncement("u1")
	if err != nil || got.ID != id || got.Content != "改期到明晚维护" {
		t.Fatalf("content change must re-notify: %v %+v", err, got)
	}
	// no-op content resave: stays read after ack
	if serr := svc.AdminUpdateAnnouncement(id, "维护通知", "改期到明晚维护", nil, "test"); serr != nil {
		t.Fatalf("resave: %+v", serr)
	}
	if err := st.AckAnnouncement(id, "u1", ""); err != nil {
		t.Fatalf("re-ack: %v", err)
	}
	if serr := svc.AdminUpdateAnnouncement(id, "维护通知", "改期到明晚维护", nil, "test"); serr != nil {
		t.Fatalf("no-op resave: %+v", serr)
	}
	if _, err := st.CurrentUnreadAnnouncement("u1"); err != store.ErrNotFound {
		t.Fatalf("no-op resave must keep read state, got %v", err)
	}
}

func boolPtr(b bool) *bool { return &b }

func TestMembershipValid(t *testing.T) {
	future := model.User{MemberExpireAt: "2099-01-01T00:00:00.000Z"}
	past := model.User{MemberExpireAt: "2020-01-01T00:00:00.000Z"}
	empty := model.User{}
	if !MembershipValid(&future) {
		t.Fatal("future must be valid")
	}
	if MembershipValid(&past) || MembershipValid(&empty) {
		t.Fatal("past/empty must be invalid")
	}
}
