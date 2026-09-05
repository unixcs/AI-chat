package service

import (
	"strings"
	"testing"

	"ai-chat-backend/internal/model"
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
	// Node/dayjs on the UTC containers parses naive strings as UTC; keep that.
	cases := map[string]string{
		"2027-01-01 00:00": "2027-01-01T00:00:00.000Z",
		"":                 "",
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
