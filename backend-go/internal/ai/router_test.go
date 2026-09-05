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

// R1 fix: after the first token reached the client, a dying entry must NOT
// replay the remaining pool (client would see "AAABBB" concatenation).
func TestNoReplayAfterFirstToken(t *testing.T) {
	// A streams partial content, then its connection dies mid-stream.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher := w.(http.Flusher)
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, `data: {"choices":[{"delta":{"content":"AAA"}}]}`+"\n\n")
		flusher.Flush()
		// hijack and kill the connection abruptly
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Fatal("cannot hijack")
		}
		conn, _, err := hj.Hijack()
		if err != nil {
			t.Fatal(err)
		}
		conn.Close()
	}))
	official := fakeUpstream(t, 0, chunks, http.StatusOK)
	defer srv.Close()
	defer official.Close()

	pool := []config.ProviderEntry{entry("partial-dies", srv.URL, false), entry("official", official.URL, true)}
	r := NewRouter(testCfg(pool, "gateway"))

	var collected strings.Builder
	_, merr := r.Stream(context.Background(), ChatRequest{Messages: []ChatMessage{{Role: "user", Content: "hi"}}},
		func(delta string) { collected.WriteString(delta) }, nil)
	if merr == nil {
		t.Fatal("mid-stream death must surface an error")
	}
	if merr.Code != "STREAM_ERROR" {
		t.Fatalf("expected STREAM_ERROR, got %s (%s)", merr.Code, merr.Message)
	}
	if collected.String() != "AAA" {
		t.Fatalf("client must receive exactly the partial content, got %q", collected.String())
	}
	if snap := r.HealthSnapshot(); snap[1].Successes != 0 {
		t.Fatalf("official must NOT be used after window closed, got %d successes", snap[1].Successes)
	}
}

// R1 fix: fallback entry must be exempt from the 2s first-token watchdog —
// an official API that starts streaming at 3s still serves (final safety net).
func TestFallbackExemptFromFirstTokenWatchdog(t *testing.T) {
	slowPool := fakeUpstream(t, 2*time.Second, chunks, http.StatusOK)  // burns itself, timeout at 400ms
	lateOfficial := fakeUpstream(t, 3*time.Second, chunks, http.StatusOK) // first token at 3s > 2s watchdog
	defer slowPool.Close()
	defer lateOfficial.Close()

	pool := []config.ProviderEntry{
		entry("slowPool", slowPool.URL, false),
		entry("official", lateOfficial.URL, true),
	}
	cfg := testCfg(pool, "gateway")
	cfg.AIGatewayBudgetMs = 300 // exhausted → straight to official
	r := NewRouter(cfg)

	var collected strings.Builder
	result, merr := r.Stream(context.Background(), ChatRequest{Messages: []ChatMessage{{Role: "user", Content: "hi"}}},
		func(delta string) { collected.WriteString(delta) }, nil)
	if merr != nil {
		t.Fatalf("official fallback must not be cut by the first-token watchdog: %+v", merr)
	}
	if result.EntryName != "official" || collected.String() != "你好世界" {
		t.Fatalf("unexpected result %+v %q", result, collected.String())
	}
}

// R1 fix: client cancel must not poison health (no failures, no cooldowns,
// no further entries tried) — stop-generation is not an upstream fault.
func TestClientCancelNoHealthDamage(t *testing.T) {
	dead := "http://127.0.0.1:1"
	healthy := fakeUpstream(t, 0, chunks, http.StatusOK)
	defer healthy.Close()

	pool := []config.ProviderEntry{
		entry("dead", dead, false),
		entry("healthy", healthy.URL, false),
		entry("official", healthy.URL, true),
	}
	cfg := testCfg(pool, "gateway")
	r := NewRouter(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond) // let the first (dead) attempt start
		cancel()
	}()
	_, merr := r.Stream(ctx, ChatRequest{Messages: []ChatMessage{{Role: "user", Content: "hi"}}}, nil, nil)
	if merr == nil {
		t.Fatal("cancelled request must return an error")
	}
	for _, snap := range r.HealthSnapshot() {
		if snap.Failures != 0 || snap.ConsecFails != 0 || snap.CooldownSecs != 0 {
			t.Fatalf("client cancel must not damage health: %+v", snap)
		}
	}
}

// Queue overflow surfaces the Node-compatible code/message. White-box: fill
// the in-flight semaphore and the waiting queue, then the next Stream call
// must be rejected without touching any upstream.
func TestQueueOverflow(t *testing.T) {
	cfg := testCfg([]config.ProviderEntry{entry("never-called", "http://127.0.0.1:1", false)}, "gateway")
	cfg.ModelConcurrency = 1
	cfg.ModelQueueMax = 1
	r := NewRouter(cfg)

	// occupy the single in-flight slot and the single waiting slot
	r.sem <- struct{}{}
	r.queued <- struct{}{}

	_, merr := r.Stream(context.Background(), ChatRequest{Messages: []ChatMessage{{Role: "user", Content: "hi"}}}, nil, nil)
	if merr == nil || merr.Code != "MODEL_QUEUE_OVERFLOW" {
		t.Fatalf("expected MODEL_QUEUE_OVERFLOW, got %+v", merr)
	}
	if merr.Message != "当前请求较多，请稍后再试" {
		t.Fatalf("message must match Node contract, got %q", merr.Message)
	}
	// stats must not count a rejected request as attempted work
	if snap := r.StatsSnapshot(); snap["totalRequests"] != 1 {
		t.Fatalf("totalRequests accounting: %v", snap)
	}
}

// Official 402 (insufficient balance) terminates immediately — switching
// entries cannot fix an account-level problem.
func TestOfficial402StopsImmediately(t *testing.T) {
	official := fakeUpstream(t, 0, chunks, http.StatusPaymentRequired)
	defer official.Close()

	p := []config.ProviderEntry{entry("dead", "http://127.0.0.1:1", false), entry("official", official.URL, true)}
	r := NewRouter(testCfg(p, "gateway"))
	_, merr := r.Stream(context.Background(), ChatRequest{Messages: []ChatMessage{{Role: "user", Content: "hi"}}}, nil, nil)
	if merr == nil || merr.Code != "MODEL_INSUFFICIENT_BALANCE" {
		t.Fatalf("expected MODEL_INSUFFICIENT_BALANCE, got %+v", merr)
	}
}
