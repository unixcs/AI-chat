package service

import (
	"path/filepath"
	"strings"
	"testing"

	"ai-chat-backend/internal/model"
	"ai-chat-backend/internal/store"
)

// P3: system prompt = base + 长度 + 风格 + 格式, one labeled segment per
// dimension in that order — including the standard options (declared
// semantic change: defaults now carry their card text too).
func TestBuildSystemPromptCards(t *testing.T) {
	svc := promptTestService(t)

	base := svc.EffectiveSystemPrompt()
	u1 := svc.BuildSystemPrompt(&model.User{ID: "u1", AnswerLength: "concise", AnswerStyle: "plain", AnswerFormat: "plain"})
	u2 := svc.BuildSystemPrompt(&model.User{ID: "u2"})
	u3 := svc.BuildSystemPrompt(&model.User{ID: "u3", AnswerLength: "", AnswerStyle: "standard"})

	if !strings.HasPrefix(u1, base) {
		t.Fatalf("base prompt must come first: %q", u1)
	}
	seg := func(s string) []string {
		return strings.Split(s, "\n\n")
	}
	parts := seg(u1)
	if len(parts) != 4 {
		t.Fatalf("expected base+3 segments, got %d: %q", len(parts), u1)
	}
	if !strings.HasPrefix(parts[1], "回答长度要求：") || !strings.Contains(parts[1], "精简") {
		t.Fatalf("length segment wrong: %q", parts[1])
	}
	if !strings.HasPrefix(parts[2], "回答风格要求：") || !strings.Contains(parts[2], "大白话") {
		t.Fatalf("style segment wrong: %q", parts[2])
	}
	if !strings.HasPrefix(parts[3], "输出格式要求：") || !strings.Contains(parts[3], "纯文字") {
		t.Fatalf("format segment wrong: %q", parts[3])
	}

	// defaults (empty values) normalize to the standard cards — which are
	// non-empty presets, so a default user gets three standard segments too
	parts2 := seg(u2)
	if len(parts2) != 4 || !strings.Contains(parts2[1], "适中") || !strings.Contains(parts2[3], "排版方式自由") {
		t.Fatalf("default user must get standard cards: %q", u2)
	}
	parts3 := seg(u3)
	if !strings.Contains(parts3[1], "适中") || !strings.Contains(parts3[2], "自然、清晰、友好") {
		t.Fatalf("partial prefs must fall back per-dimension: %q", u3)
	}
}

// P3b: an admin-blanked card contributes no segment.
func TestBuildSystemPromptEmptyCardSkipsSegment(t *testing.T) {
	svc := promptTestService(t)
	if serr := svc.AdminUpdatePrefPrompt("answerFormat", "plain", "", "tester"); serr != nil {
		t.Fatalf("blank card: %+v", serr)
	}
	got := svc.BuildSystemPrompt(&model.User{AnswerFormat: "plain"})
	if strings.Contains(got, "输出格式要求") {
		t.Fatalf("empty card must skip its segment: %q", got)
	}
	if !strings.Contains(got, "回答长度要求") {
		t.Fatalf("other dimensions unaffected: %q", got)
	}
}

// P4: seeding is idempotent and never overwrites admin edits across restarts.
func TestPrefCardSeedIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "t.sqlite")
	st, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.SetPrefPromptContent("answerLength", "concise", "管理员改过的话"); err != nil {
		t.Fatalf("admin edit: %v", err)
	}
	st.DB.Close()

	st2, err := store.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer st2.DB.Close()
	// second explicit seeding must be a no-op
	if err := st2.SeedPrefPrompts(); err != nil {
		t.Fatal(err)
	}
	cards, err := st2.ListPrefPrompts()
	if err != nil || len(cards) != 10 {
		t.Fatalf("10 cards expected, got %d err=%v", len(cards), err)
	}
	for _, c := range cards {
		if c.Dimension == "answerLength" && c.Value == "concise" {
			if c.Content != "管理员改过的话" || !c.Customized {
				t.Fatalf("admin edit lost on reseed: %+v", c)
			}
		}
	}
}

// P7: reset drops the override and falls back to the factory preset.
func TestPrefCardReset(t *testing.T) {
	svc := promptTestService(t)
	before, _ := svc.Store.GetPrefPromptContent("answerStyle", "rigorous")
	if serr := svc.AdminUpdatePrefPrompt("answerStyle", "rigorous", "自定义严谨", "tester"); serr != nil {
		t.Fatalf("update: %+v", serr)
	}
	if got, _ := svc.Store.GetPrefPromptContent("answerStyle", "rigorous"); got != "自定义严谨" {
		t.Fatalf("edit not visible: %q", got)
	}
	if serr := svc.AdminResetPrefPrompt("answerStyle", "rigorous", "tester"); serr != nil {
		t.Fatalf("reset: %+v", serr)
	}
	after, _ := svc.Store.GetPrefPromptContent("answerStyle", "rigorous")
	if after != before {
		t.Fatalf("reset must restore preset: %q != %q", after, before)
	}
}

// P8: users created before the format dimension exist keep a NULL column —
// SafeUser reports null and callers normalize to standard.
func TestLegacyUserAnswerFormatNull(t *testing.T) {
	svc := promptTestService(t)
	svc.Store.CreateUser(&model.User{ID: "old", Phone: "13900000009", Status: "active", Role: "user"})
	u, err := svc.Store.GetUserByID("old")
	if err != nil {
		t.Fatal(err)
	}
	if u.AnswerFormat != "" {
		t.Fatalf("legacy user must have empty answerFormat, got %q", u.AnswerFormat)
	}
	if got := u.SafeUser()["answerFormat"]; got != nil {
		t.Fatalf("SafeUser must report null answerFormat for legacy users, got %v", got)
	}
	// empty value normalizes to the standard card
	if got := svc.GetPrefPromptContent("answerFormat", ""); !strings.Contains(got, "排版方式自由") {
		t.Fatalf("empty value must resolve to standard card: %q", got)
	}
}

// P9 (service level): absent/empty fields leave stored preferences untouched.
func TestUpdatePreferencesPartial(t *testing.T) {
	svc := promptTestService(t)
	svc.Store.CreateUser(&model.User{ID: "u", Phone: "13900000010", Status: "active", Role: "user"})
	if serr := svc.UpdatePreferences("u", "concise", "plain", "plain"); serr != nil {
		t.Fatalf("full update: %+v", serr)
	}
	// only length provided → style/format untouched
	if serr := svc.UpdatePreferences("u", "detailed", "", ""); serr != nil {
		t.Fatalf("partial update: %+v", serr)
	}
	u, _ := svc.Store.GetUserByID("u")
	if u.AnswerLength != "detailed" || u.AnswerStyle != "plain" || u.AnswerFormat != "plain" {
		t.Fatalf("partial update must not clear others: %+v", u)
	}
	// invalid format rejected
	if serr := svc.UpdatePreferences("u", "", "", "markdown"); serr == nil || serr.Status != 400 {
		t.Fatalf("invalid format must 400, got %+v", serr)
	}
}

// P6 (service level): unknown cards are rejected.
func TestAdminUpdatePrefPromptUnknownCard(t *testing.T) {
	svc := promptTestService(t)
	if serr := svc.AdminUpdatePrefPrompt("answerTone", "warm", "x", "tester"); serr == nil || serr.Status != 404 {
		t.Fatalf("unknown dimension must 404, got %+v", serr)
	}
	if serr := svc.AdminUpdatePrefPrompt("answerLength", "tiny", "x", "tester"); serr == nil || serr.Status != 404 {
		t.Fatalf("unknown value must 404, got %+v", serr)
	}
}
