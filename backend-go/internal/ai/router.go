// Package ai isolates every upstream-model concern from the business layer:
// provider entries (OpenAI-compatible), the health tracker, and the failover
// router. Business code calls Router.Stream and never learns which model won.
package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"ai-chat-backend/internal/config"
)

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Messages []ChatMessage
}

// StreamResult reports which entry served the request (for stats/logs only).
type StreamResult struct {
	EntryName    string
	Model        string
	FirstTokenMs int64
	TotalMs      int64
	ContentChars int
	FallbackUsed bool // served by the official-API backstop
	Switches     int  // how many entries failed/were skipped before success
}

type ModelError struct {
	Code    string // MODEL_TIMEOUT / MODEL_RATE_LIMIT / MODEL_UPSTREAM_ERROR / MODEL_INSUFFICIENT_BALANCE / MODEL_QUEUE_OVERFLOW
	Message string
}

func (e *ModelError) Error() string { return e.Message }

// ---------- health tracker ----------

type entryHealth struct {
	ConsecFails    int64
	LastErr        string
	CooldownUntil  time.Time
	Successes      int64
	Failures       int64
	Timeouts       int64
	LastFirstToken int64
	EwmaFirstToken float64
	EwmaTotal      float64
	LastFailAt     time.Time
}

type Health struct {
	mu      sync.Mutex
	byEntry map[string]*entryHealth
	baseCD  time.Duration
	maxCD   time.Duration
}

func NewHealth() *Health {
	return &Health{
		byEntry: map[string]*entryHealth{},
		baseCD:  30 * time.Second,
		maxCD:   10 * time.Minute,
	}
}

func (h *Health) get(name string) *entryHealth {
	e, ok := h.byEntry[name]
	if !ok {
		e = &entryHealth{}
		h.byEntry[name] = e
	}
	return e
}

func (h *Health) Available(name string, now time.Time) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return now.After(h.get(name).CooldownUntil)
}

func (h *Health) RecordSuccess(name string, firstTokenMs, totalMs int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	e := h.get(name)
	e.ConsecFails = 0
	e.LastErr = ""
	e.CooldownUntil = time.Time{}
	e.Successes++
	e.LastFirstToken = firstTokenMs
	if e.EwmaFirstToken == 0 {
		e.EwmaFirstToken = float64(firstTokenMs)
	} else {
		e.EwmaFirstToken = e.EwmaFirstToken*0.7 + float64(firstTokenMs)*0.3
	}
	if e.EwmaTotal == 0 {
		e.EwmaTotal = float64(totalMs)
	} else {
		e.EwmaTotal = e.EwmaTotal*0.7 + float64(totalMs)*0.3
	}
}

// incidentWindow groups near-simultaneous failures into ONE incident: a burst
// of 50 concurrent requests hitting a dead upstream must count as a single
// failure for escalation purposes, otherwise one blip puts the whole pool on
// minutes-long cooldown. Escalation only happens between incidents.
const incidentWindow = 10 * time.Second

func (h *Health) RecordFailure(name string, isTimeout bool, msg string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	e := h.get(name)
	e.Failures++
	now := time.Now()
	if e.LastFailAt.IsZero() || now.Sub(e.LastFailAt) > incidentWindow {
		e.ConsecFails++
	}
	e.LastFailAt = now
	if isTimeout {
		e.Timeouts++
	}
	e.LastErr = msg
	// Exponential cooldown per incident: 30s, 1m, 2m, 4m ... capped. Once the
	// cooldown expires the entry is simply eligible again — its next natural
	// attempt doubles as the recovery probe, so no separate prober is needed.
	cd := h.baseCD << uint(min64(e.ConsecFails-1, 5))
	if cd > h.maxCD {
		cd = h.maxCD
	}
	e.CooldownUntil = now.Add(cd)
}

func min64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

type HealthSnapshot struct {
	Name           string  `json:"name"`
	Model          string  `json:"model"`
	Available      bool    `json:"available"`
	ConsecFails    int64   `json:"consecutiveFails"`
	Successes      int64   `json:"successes"`
	Failures       int64   `json:"failures"`
	Timeouts       int64   `json:"timeouts"`
	LastFirstToken int64   `json:"lastFirstTokenMs"`
	EwmaFirstToken float64 `json:"ewmaFirstTokenMs"`
	EwmaTotal      float64 `json:"ewmaTotalMs"`
	LastErr        string  `json:"lastError,omitempty"`
	CooldownSecs   int64   `json:"cooldownRemainingSecs"`
}

// ---------- stats ----------

type Stats struct {
	mu               sync.Mutex
	gatewayRequests  int64
	gatewaySuccesses int64
	officialFallback int64
	switches         int64
	totalRequests    int64
	totalSuccesses   int64
}

func (s *Stats) RecordRequest() {
	s.mu.Lock()
	s.totalRequests++
	s.mu.Unlock()
}

// RecordGatewayAttempt counts every real (non-skipped) gateway entry attempt.
func (s *Stats) RecordGatewayAttempt() {
	s.mu.Lock()
	s.gatewayRequests++
	s.mu.Unlock()
}

func (s *Stats) RecordOutcome(isFallback bool, switches int) {
	s.mu.Lock()
	s.totalSuccesses++
	if isFallback {
		s.officialFallback++
	} else {
		s.gatewaySuccesses++
	}
	s.switches += int64(switches)
	s.mu.Unlock()
}

func (s *Stats) Snapshot() map[string]int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return map[string]int64{
		"totalRequests":   s.totalRequests,
		"totalSuccesses":  s.totalSuccesses,
		"gatewayRequests": s.gatewayRequests,
		"gatewaySuccess":  s.gatewaySuccesses,
		"officialFallback": s.officialFallback,
		"switches":        s.switches,
	}
}

// ---------- router ----------

type Router struct {
	entries []config.ProviderEntry
	cfg     *config.Config
	client  *http.Client
	health  *Health
	stats   *Stats

	sem    chan struct{} // in-flight upstream attempts
	queued chan struct{} // waiting slots
}

func NewRouter(cfg *config.Config) *Router {
	// Config-level normalization: programmatic callers can hand us zero
	// values that would otherwise brick the pool (0ms watchdog fires at once;
	// a 0-cap semaphore blocks forever).
	if cfg.AIFirstTokenTimeoutMs <= 0 {
		cfg.AIFirstTokenTimeoutMs = 2000
	}
	if cfg.AIGatewayBudgetMs <= 0 {
		cfg.AIGatewayBudgetMs = 6000
	}
	if cfg.AITotalTimeoutMs <= 0 {
		cfg.AITotalTimeoutMs = 90000
	}
	if cfg.ModelConcurrency <= 0 {
		cfg.ModelConcurrency = 6
	}
	if cfg.ModelQueueMax <= 0 {
		cfg.ModelQueueMax = 50
	}
	entries := make([]config.ProviderEntry, 0, len(cfg.AIProviders))
	for _, e := range cfg.AIProviders {
		// Official mode serves ONLY via the official backstop — the runtime
		// enforces this even if the config handed us a full pool.
		if cfg.AIMode == "official" && !e.Fallback {
			continue
		}
		if e.FirstTokenTimeoutMs <= 0 {
			e.FirstTokenTimeoutMs = cfg.AIFirstTokenTimeoutMs
		}
		entries = append(entries, e)
	}
	return &Router{
		entries: entries,
		cfg:     cfg,
		client:  &http.Client{Timeout: 0}, // deadlines enforced via per-attempt context
		health:  NewHealth(),
		stats:   &Stats{},
		sem:     make(chan struct{}, cfg.ModelConcurrency),
		queued:  make(chan struct{}, cfg.ModelQueueMax),
	}
}

func (r *Router) HealthSnapshot() []HealthSnapshot {
	now := time.Now()
	out := make([]HealthSnapshot, 0, len(r.entries))
	for _, e := range r.entries {
		r.health.mu.Lock()
		h := r.health.get(e.Name)
		snap := HealthSnapshot{
			Name:           e.Name,
			Model:          e.Model,
			ConsecFails:    h.ConsecFails,
			Successes:      h.Successes,
			Failures:       h.Failures,
			Timeouts:       h.Timeouts,
			LastFirstToken: h.LastFirstToken,
			EwmaFirstToken: h.EwmaFirstToken,
			EwmaTotal:      h.EwmaTotal,
			LastErr:        h.LastErr,
			Available:      now.After(h.CooldownUntil),
		}
		if !h.CooldownUntil.IsZero() && now.Before(h.CooldownUntil) {
			snap.CooldownSecs = int64(h.CooldownUntil.Sub(now).Seconds())
		}
		r.health.mu.Unlock()
		out = append(out, snap)
	}
	return out
}

func (r *Router) StatsSnapshot() map[string]int64 { return r.stats.Snapshot() }
func (r *Router) Mode() string                    { return r.cfg.AIMode }
func (r *Router) HasAPIKey() bool                 { return r.cfg.AIAPIKey != "" }

// Stream walks the pool in order. Every gateway entry gets a first-token
// budget (~2s); once the cumulative gateway budget (~6s) is burnt, the official
// DeepSeek entry runs with a full-length budget (no first-token watchdog — the
// official API is the final safety net and must never be cut by the 2s rule).
// Once a first token reaches the client the failover window is closed: a
// mid-stream failure surfaces as an error and MUST NOT replay remaining entries
// (that would concatenate duplicated text).
func (r *Router) Stream(ctx context.Context, req ChatRequest, onDelta func(string), onModel func(string)) (*StreamResult, *ModelError) {
	r.stats.RecordRequest()

	if err := r.acquireSlot(ctx); err != nil {
		switch {
		case errors.Is(err, errQueueFull):
			return nil, &ModelError{Code: "MODEL_QUEUE_OVERFLOW", Message: "当前请求较多，请稍后再试"}
		case errors.Is(ctx.Err(), context.Canceled):
			return nil, &ModelError{Code: "MODEL_CLIENT_GONE", Message: "客户端已取消"}
		default:
			return nil, &ModelError{Code: "MODEL_TIMEOUT", Message: "模型响应超时，请重试"}
		}
	}
	defer func() { <-r.sem }()

	// The gateway budget measures MODEL time, not queue time: the clock starts
	// only once this request owns a slot. (Queue waits used to burn the budget
	// and silently route every queued request to the paid official API.)
	start := time.Now()

	// emitted tracks whether any content token already reached the client.
	// After that point the failover window is closed — see the loop below.
	emitted := false
	clientDelta := func(delta string) {
		emitted = true
		if onDelta != nil {
			onDelta(delta)
		}
	}

	gatewayBudgetUntil := start.Add(time.Duration(r.cfg.AIGatewayBudgetMs) * time.Millisecond)
	var switches int
	var lastErr *ModelError

	for _, entry := range r.entries {
		isFallback := entry.Fallback
		if isFallback && !r.HasAPIKey() {
			break
		}

		// Client is gone (stop generation / closed tab): stop everything.
		// No further entries are tried and no health damage is recorded —
		// a user-initiated cancel says nothing about upstream health.
		if ctx.Err() != nil {
			return nil, &ModelError{Code: "MODEL_CLIENT_GONE", Message: "客户端已取消"}
		}

		// Skip cooling-down entries unless they are the last resort.
		if !isFallback && !r.health.Available(entry.Name, time.Now()) {
			continue
		}

		// Gateway entries share a cumulative budget; once burnt, jump to the
		// official backstop. The user must never be left waiting on dead models.
		if !isFallback && time.Now().After(gatewayBudgetUntil) {
			switches++
			continue
		}

		if !isFallback {
			r.stats.RecordGatewayAttempt()
		}

		result, merr := r.streamEntry(ctx, entry, req, clientDelta, onModel)
		if merr == nil {
			r.health.RecordSuccess(entry.Name, result.FirstTokenMs, result.TotalMs)
			r.stats.RecordOutcome(isFallback, switches)
			result.Switches = switches
			return result, nil
		}

		if merr.Code == "MODEL_CLIENT_GONE" || ctx.Err() != nil {
			return nil, merr
		}

		// A user-initiated cancel is not an upstream fault — no health damage.
		if merr.Code != "MODEL_CLIENT_GONE" {
			isTimeout := merr.Code == "MODEL_TIMEOUT"
			r.health.RecordFailure(entry.Name, isTimeout, merr.Message)
			switches++
			logAttempt(entry.Name, entry.Model, merr.Code, merr.Message)
		}

		// Failover window is closed once the client saw content: surface the
		// failure immediately instead of replaying the pool (no "AAABBB").
		// The failing entry still gets its health record above — a model that
		// always dies mid-stream must not keep winning the first window.
		if emitted {
			return nil, &ModelError{Code: "STREAM_ERROR", Message: "模型响应中断"}
		}

		lastErr = merr

		// Balance problems on the official account will not improve by
		// switching — surface immediately instead of burning the pool.
		if merr.Code == "MODEL_INSUFFICIENT_BALANCE" && isFallback {
			return nil, merr
		}
	}

	if lastErr == nil {
		lastErr = &ModelError{Code: "MODEL_UPSTREAM_ERROR", Message: "模型调用失败"}
	}
	return nil, lastErr
}

// errQueueFull marks a rejected request: every in-flight slot is busy and the
// waiting queue has hit ModelQueueMax (Node-compatible reject semantics —
// waiting forever would let slow clients pin memory during an upstream outage).
var errQueueFull = errors.New("model queue full")

func (r *Router) acquireSlot(ctx context.Context) error {
	select {
	case r.queued <- struct{}{}:
		defer func() { <-r.queued }()
		select {
		case r.sem <- struct{}{}:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	default:
		return errQueueFull
	}
}

// streamEntry runs one upstream attempt. Non-fallback entries are bound by a
// first-token watchdog; the fallback (official API) entry is exempt — it is the
// final safety net and must never be cut by the 2s rule (its budget is the
// overall total timeout).
func (r *Router) streamEntry(ctx context.Context, entry config.ProviderEntry, req ChatRequest, onDelta func(string), onModel func(name string)) (*StreamResult, *ModelError) {
	attemptStart := time.Now()
	attemptCtx, cancel := context.WithTimeout(ctx, time.Duration(r.cfg.AITotalTimeoutMs)*time.Millisecond)
	defer cancel()

	var watchdogFired atomic.Bool
	var watchdog *time.Timer
	if !entry.Fallback {
		watchdog = time.AfterFunc(time.Duration(entry.FirstTokenTimeoutMs)*time.Millisecond, func() {
			watchdogFired.Store(true)
			cancel() // nothing streamed yet — kill the attempt so failover can proceed
		})
		defer watchdog.Stop()
	}

	body := map[string]any{
		"model":    entry.Model,
		"stream":   true,
		"messages": req.Messages,
	}
	if entry.ExtraBodyRaw != "" {
		var extra map[string]any
		if err := json.Unmarshal([]byte(entry.ExtraBodyRaw), &extra); err == nil {
			for k, v := range extra {
				body[k] = v
			}
		}
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, &ModelError{Code: "MODEL_UPSTREAM_ERROR", Message: "模型请求构造失败"}
	}

	httpReq, err := http.NewRequestWithContext(attemptCtx, http.MethodPost, entry.BaseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return nil, &ModelError{Code: "MODEL_UPSTREAM_ERROR", Message: "模型请求构造失败"}
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if entry.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+entry.APIKey)
	}

	resp, err := r.client.Do(httpReq)
	if err != nil {
		if watchdogFired.Load() {
			return nil, &ModelError{Code: "MODEL_TIMEOUT", Message: "模型首字响应超时"}
		}
		if errors.Is(ctx.Err(), context.Canceled) {
			return nil, &ModelError{Code: "MODEL_CLIENT_GONE", Message: "客户端已取消"}
		}
		if errors.Is(attemptCtx.Err(), context.DeadlineExceeded) {
			return nil, &ModelError{Code: "MODEL_TIMEOUT", Message: "模型响应超时，请重试"}
		}
		return nil, &ModelError{Code: "MODEL_UPSTREAM_ERROR", Message: "模型服务异常: " + err.Error()}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		switch {
		case resp.StatusCode == 402:
			return nil, &ModelError{Code: "MODEL_INSUFFICIENT_BALANCE", Message: "模型余额不足，请联系管理员"}
		case resp.StatusCode == 429:
			return nil, &ModelError{Code: "MODEL_RATE_LIMIT", Message: "模型请求过于频繁，请稍后重试"}
		case resp.StatusCode >= 500:
			return nil, &ModelError{Code: "MODEL_UPSTREAM_ERROR", Message: fmt.Sprintf("模型服务异常: %d", resp.StatusCode)}
		default:
			return nil, &ModelError{Code: "MODEL_UPSTREAM_ERROR", Message: fmt.Sprintf("模型请求失败: %d %s", resp.StatusCode, truncate(string(detail), 200))}
		}
	}

	if onModel != nil {
		onModel(entry.Name)
	}

	reader := bufio.NewReaderSize(resp.Body, 64*1024)
	var firstToken int64
	var gotFirstToken bool
	var chars int

	for {
		line, err := readLine(reader)
		if err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				// Client-initiated cancel is classified first — even after the
				// first token — so health and error contracts stay honest.
				return nil, &ModelError{Code: "MODEL_CLIENT_GONE", Message: "客户端已取消"}
			}
			if watchdogFired.Load() {
				return nil, &ModelError{Code: "MODEL_TIMEOUT", Message: "模型首字响应超时"}
			}
			if errors.Is(attemptCtx.Err(), context.DeadlineExceeded) {
				// Total-timeout death counts as a timeout regardless of whether
				// content had already started flowing.
				return nil, &ModelError{Code: "MODEL_TIMEOUT", Message: "模型响应超时，请重试"}
			}
			if gotFirstToken {
				// Upstream died mid-stream; the client already has content.
				return nil, &ModelError{Code: "MODEL_UPSTREAM_ERROR", Message: "模型响应中断"}
			}
			return nil, &ModelError{Code: "MODEL_UPSTREAM_ERROR", Message: "模型服务异常: " + err.Error()}
		}

		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if debugAI() {
			fmt.Printf("[ai-debug] entry=%s line=%q\n", entry.Name, line)
		}
		if data == "[DONE]" {
			break
		}

		var chunk struct {
			Choices []struct {
				Delta struct {
					Content          string `json:"content"`
					ReasoningContent string `json:"reasoning_content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if len(chunk.Choices) == 0 {
			continue
		}
		// reasoning_content is billed upstream but never shown; forward only
		// visible content (same policy as the Node backend).
		delta := chunk.Choices[0].Delta.Content
		if delta == "" {
			continue
		}
		if !gotFirstToken {
			firstToken = time.Since(attemptStart).Milliseconds()
			gotFirstToken = true
			if watchdog != nil {
				watchdog.Stop()
			}
		}
		chars += len(delta)
		if onDelta != nil {
			onDelta(delta)
		}
	}

	if !gotFirstToken {
		return nil, &ModelError{Code: "MODEL_UPSTREAM_ERROR", Message: "模型返回空响应"}
	}

	return &StreamResult{
		EntryName:    entry.Name,
		Model:        entry.Model,
		FirstTokenMs: firstToken,
		TotalMs:      time.Since(attemptStart).Milliseconds(),
		ContentChars: chars,
		FallbackUsed: entry.Fallback,
	}, nil
}

// readLine reads one line with a size cap so a hostile upstream cannot exhaust memory.
func readLine(reader *bufio.Reader) (string, error) {
	var sb strings.Builder
	for {
		chunk, err := reader.ReadSlice('\n')
		sb.Write(chunk)
		if err == nil {
			return sb.String(), nil
		}
		if errors.Is(err, bufio.ErrBufferFull) {
			if sb.Len() > 1<<20 {
				return "", errors.New("sse line too long")
			}
			continue
		}
		return sb.String(), err
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func logAttempt(name, model, code, msg string) {
	fmt.Printf("[ai-router] entry=%s model=%s outcome=%s detail=%q\n", name, model, code, msg)
}

func debugAI() bool { return os.Getenv("AI_DEBUG") == "1" }
