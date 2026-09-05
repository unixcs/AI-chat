package ai

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ai-chat-backend/internal/config"
)

// fakeUpstream builds an OpenAI-compatible SSE server that emits the given
// content chunks after an optional delay. delay > firstTokenTimeout must
// trigger failover.
func fakeUpstream(t *testing.T, firstChunkDelay time.Duration, chunks []string, statusCode int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if statusCode != http.StatusOK {
			w.WriteHeader(statusCode)
			_, _ = w.Write([]byte(`{"error":"boom"}`))
			return
		}
		flusher := w.(http.Flusher)
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher.Flush()
		if firstChunkDelay > 0 {
			time.Sleep(firstChunkDelay)
		}
		for _, c := range chunks {
			fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"content\":%q}}]}\n\n", c)
			flusher.Flush()
		}
		fmt.Fprint(w, "data: [DONE]\n\n")
		flusher.Flush()
	}))
}

func testCfg(entries []config.ProviderEntry, mode string) *config.Config {
	return &config.Config{
		AIMode:                mode,
		AIAPIKey:              "sk-test",
		AIFirstTokenTimeoutMs: 400,
		AIGatewayBudgetMs:     1200,
		AITotalTimeoutMs:      5000,
		ModelConcurrency:      4,
		ModelQueueMax:         10,
		AIProviders:           entries,
	}
}

func entry(name, url string, fallback bool) config.ProviderEntry {
	return config.ProviderEntry{
		Name:     name,
		BaseURL:  url,
		APIKey:   "sk-upstream",
		Model:    "test-model",
		Fallback: fallback,
	}
}

var chunks = []string{"你", "好", "世", "界"}

func TestRouterFailoverFromSlowToFast(t *testing.T) {
	slow := fakeUpstream(t, 2*time.Second, chunks, http.StatusOK)   // exceeds first-token budget
	dead := "http://127.0.0.1:1"                                    // connection refused
	fast := fakeUpstream(t, 0, chunks, http.StatusOK)
	official := fakeUpstream(t, 0, chunks, http.StatusOK)
	defer slow.Close()
	defer fast.Close()
	defer official.Close()

	pool := []config.ProviderEntry{
		entry("slow", slow.URL, false),
		entry("dead", dead, false),
		entry("fast", fast.URL, false),
		entry("official", official.URL, true),
	}
	r := NewRouter(testCfg(pool, "gateway"))

	var collected strings.Builder
	result, merr := r.Stream(context.Background(), ChatRequest{Messages: []ChatMessage{{Role: "user", Content: "hi"}}},
		func(delta string) { collected.WriteString(delta) }, nil)
	if merr != nil {
		t.Fatalf("expected success, got %+v", merr)
	}
	if result.EntryName != "fast" {
		t.Fatalf("expected fast to serve, got %s", result.EntryName)
	}
	if result.FallbackUsed {
		t.Fatal("fallback must not be used when a gateway entry works")
	}
	if result.Switches != 2 {
		t.Fatalf("expected 2 switches (slow+dead), got %d", result.Switches)
	}
	if collected.String() != "你好世界" {
		t.Fatalf("unexpected content %q", collected.String())
	}
}

// The whole gateway pool dies → official backstop must take over, and the
// client still receives the full answer without resending.
func TestRouterOfficialFallbackWhenPoolDead(t *testing.T) {
	official := fakeUpstream(t, 0, chunks, http.StatusOK)
	defer official.Close()

	pool := []config.ProviderEntry{
		entry("dead1", "http://127.0.0.1:1", false),
		entry("dead2", "http://127.0.0.1:2", false),
		entry("official", official.URL, true),
	}
	r := NewRouter(testCfg(pool, "gateway"))

	var collected strings.Builder
	result, merr := r.Stream(context.Background(), ChatRequest{Messages: []ChatMessage{{Role: "user", Content: "hi"}}},
		func(delta string) { collected.WriteString(delta) }, nil)
	if merr != nil {
		t.Fatalf("expected official fallback success, got %+v", merr)
	}
	if !result.FallbackUsed || result.EntryName != "official" {
		t.Fatalf("expected official fallback, got %s", result.EntryName)
	}
	if collected.String() != "你好世界" {
		t.Fatalf("unexpected content %q", collected.String())
	}
}

// Official mode must ONLY use the official entry, even when gateway entries exist.
func TestRouterOfficialModeIgnoresPool(t *testing.T) {
	poolEntry := fakeUpstream(t, 0, chunks, http.StatusOK)
	official := fakeUpstream(t, 0, chunks, http.StatusOK)
	defer poolEntry.Close()
	defer official.Close()

	pool := []config.ProviderEntry{
		entry("pool", poolEntry.URL, false),
		entry("official", official.URL, true),
	}
	r := NewRouter(testCfg(pool, "official"))

	result, merr := r.Stream(context.Background(), ChatRequest{Messages: []ChatMessage{{Role: "user", Content: "hi"}}}, nil, nil)
	if merr != nil {
		t.Fatalf("unexpected error %+v", merr)
	}
	if result.EntryName != "official" {
		t.Fatalf("official mode must serve via official entry, got %s", result.EntryName)
	}
}

// Health: an entry that failed gets cooled down and skipped on the next call,
// then recovers after the cooldown expires.
func TestHealthCooldownAndRecovery(t *testing.T) {
	bad := fakeUpstream(t, 2*time.Second, chunks, http.StatusOK) // always times out
	good := fakeUpstream(t, 0, chunks, http.StatusOK)
	defer bad.Close()
	defer good.Close()

	pool := []config.ProviderEntry{entry("bad", bad.URL, false), entry("good", good.URL, false)}
	r := NewRouter(testCfg(pool, "gateway"))
	r.health.baseCD = 300 * time.Millisecond // shrink cooldown for the test

	if _, merr := r.Stream(context.Background(), ChatRequest{Messages: []ChatMessage{{Role: "user", Content: "hi"}}}, nil, nil); merr != nil {
		t.Fatalf("first call failed: %+v", merr)
	}
	if result := r.HealthSnapshot()[0]; result.Available {
		t.Fatal("bad entry should be cooling down after first-token timeout")
	}

	// Second call: bad is skipped, good serves directly (no extra 2s wasted).
	start := time.Now()
	result, merr := r.Stream(context.Background(), ChatRequest{Messages: []ChatMessage{{Role: "user", Content: "hi"}}}, nil, nil)
	elapsed := time.Since(start)
	if merr != nil {
		t.Fatalf("second call failed: %+v", merr)
	}
	if result.EntryName != "good" {
		t.Fatalf("cooled-down entry must be skipped, got %s", result.EntryName)
	}
	if elapsed > 900*time.Millisecond {
		t.Fatalf("cooled-down entry must be skipped instantly, took %s", elapsed)
	}

	time.Sleep(400 * time.Millisecond)
	if !r.HealthSnapshot()[0].Available {
		t.Fatal("bad entry should be available again after cooldown")
	}
}

// reasoning_content must never reach the client stream.
func TestReasoningContentFiltered(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher := w.(http.Flusher)
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"reasoning_content\":\"thinking...\"}}]}\n\n")
		flusher.Flush()
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"答案\"}}]}\n\n")
		flusher.Flush()
		fmt.Fprint(w, "data: [DONE]\n\n")
		flusher.Flush()
	}))
	defer srv.Close()

	r := NewRouter(testCfg([]config.ProviderEntry{entry("m", srv.URL, true)}, "official"))
	var collected strings.Builder
	_, merr := r.Stream(context.Background(), ChatRequest{Messages: []ChatMessage{{Role: "user", Content: "hi"}}},
		func(delta string) { collected.WriteString(delta) }, nil)
	if merr != nil {
		t.Fatalf("unexpected error %+v", merr)
	}
	if collected.String() != "答案" {
		t.Fatalf("reasoning_content leaked: %q", collected.String())
	}
}

// Client cancellation must stop the upstream work (no goroutine leak, no panic).
func TestStreamHonorsClientCancel(t *testing.T) {
	slow := fakeUpstream(t, 3*time.Second, chunks, http.StatusOK)
	official := fakeUpstream(t, 3*time.Second, chunks, http.StatusOK)
	defer slow.Close()
	defer official.Close()

	r := NewRouter(testCfg([]config.ProviderEntry{entry("slow", slow.URL, false), entry("official", official.URL, true)}, "gateway"))
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, merr := r.Stream(ctx, ChatRequest{Messages: []ChatMessage{{Role: "user", Content: "hi"}}}, nil, nil)
	if merr == nil {
		t.Fatal("expected timeout error after cancel")
	}
	if time.Since(start) > 2*time.Second {
		t.Fatal("cancel must abort promptly, not wait for upstream")
	}
}

// Cumulative gateway budget: once burnt, remaining gateway entries are skipped
// in favor of the official backstop.
func TestGatewayBudgetExhausted(t *testing.T) {
	slow := fakeUpstream(t, 2*time.Second, chunks, http.StatusOK) // burns the whole budget while timing out
	latePool := fakeUpstream(t, 0, chunks, http.StatusOK)         // healthy but no budget left
	official := fakeUpstream(t, 0, chunks, http.StatusOK)
	defer slow.Close()
	defer latePool.Close()
	defer official.Close()

	pool := []config.ProviderEntry{
		entry("slow", slow.URL, false),
		entry("latePool", latePool.URL, false),
		entry("official", official.URL, true),
	}
	cfg := testCfg(pool, "gateway")
	cfg.AIGatewayBudgetMs = 300 // less than one first-token attempt (400ms)
	r := NewRouter(cfg)

	result, merr := r.Stream(context.Background(), ChatRequest{Messages: []ChatMessage{{Role: "user", Content: "hi"}}}, nil, nil)
	if merr != nil {
		t.Fatalf("unexpected error %+v", merr)
	}
	if result.EntryName != "official" {
		t.Fatalf("budget exhausted → official must serve, got %s", result.EntryName)
	}
}

// Budget NOT exhausted: a healthy second entry must still get its chance.
func TestGatewayBudgetAllowsHealthySecondEntry(t *testing.T) {
	slow := fakeUpstream(t, 2*time.Second, chunks, http.StatusOK)
	fast := fakeUpstream(t, 0, chunks, http.StatusOK)
	official := fakeUpstream(t, 0, chunks, http.StatusOK)
	defer slow.Close()
	defer fast.Close()
	defer official.Close()

	pool := []config.ProviderEntry{
		entry("slow", slow.URL, false),
		entry("fast", fast.URL, false),
		entry("official", official.URL, true),
	}
	cfg := testCfg(pool, "gateway")
	cfg.AIGatewayBudgetMs = 1500
	r := NewRouter(cfg)

	result, merr := r.Stream(context.Background(), ChatRequest{Messages: []ChatMessage{{Role: "user", Content: "hi"}}}, nil, nil)
	if merr != nil {
		t.Fatalf("unexpected error %+v", merr)
	}
	if result.EntryName != "fast" {
		t.Fatalf("healthy second entry must serve within budget, got %s", result.EntryName)
	}
}
