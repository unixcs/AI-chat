// End-to-end API integration tests: real HTTP stack + real SQLite in a temp
// dir + a fake OpenAI-compatible upstream, exercising the exact contract the
// Vue frontend relies on.
package api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"ai-chat-backend/internal/ai"
	"ai-chat-backend/internal/config"
	"ai-chat-backend/internal/service"
	"ai-chat-backend/internal/store"
)

type env struct {
	t        *testing.T
	server   *httptest.Server
	upstream *httptest.Server
	store    *store.Store
	cfg      *config.Config
	svc      *service.Service
}

func newEnv(t *testing.T, upstreamChunks []string) *env {
	t.Helper()
	upstream := fakeOpenAI(t, upstreamChunks)

	cfg := &config.Config{
		Port:                  "0",
		SQLITEPath:            filepath.Join(t.TempDir(), "test.sqlite"),
		JWTSecret:             "test-secret",
		SystemPrompt:          "你是测试助手。",
		SystemPromptDefault:   "你是测试助手。",
		AIModel:               "test-model",
		AIBaseURL:             upstream.URL,
		AIAPIKey:              "sk-upstream",
		AIMode:                "official",
		AIFirstTokenTimeoutMs: 1000,
		AIGatewayBudgetMs:     3000,
		AITotalTimeoutMs:      10000,
		ModelConcurrency:      4,
		ModelQueueMax:         10,
		RegistrationOpen:      true,
		RedeemOpen:            true,
	}
	// official mode with the fake upstream serving as the official entry
	cfg.AIProviders = []config.ProviderEntry{{
		Name: "official", BaseURL: upstream.URL, APIKey: "sk-upstream", Model: "test-model",
		FirstTokenTimeoutMs: 1000, Fallback: true,
	}}

	st, err := store.Open(cfg.SQLITEPath)
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	t.Cleanup(func() { st.DB.Close() })

	router := ai.NewRouter(cfg)
	svc := service.New(cfg, st, router)
	handler := New(svc, router)
	server := httptest.NewServer(handler.Routes())
	t.Cleanup(server.Close)
	return &env{t: t, server: server, upstream: upstream, store: st, cfg: cfg, svc: svc}
}

func fakeOpenAI(t *testing.T, chunks []string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		flusher := w.(http.Flusher)
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher.Flush()
		for _, c := range chunks {
			fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"content\":%q}}]}\n\n", c)
			flusher.Flush()
		}
		fmt.Fprint(w, "data: [DONE]\n\n")
		flusher.Flush()
	}))
}

// ---------- minimal HTTP client helpers ----------

func (e *env) req(method, path, token string, body map[string]any) (int, map[string]any) {
	e.t.Helper()
	var reader *strings.Reader
	if body != nil {
		raw, _ := json.Marshal(body)
		reader = strings.NewReader(string(raw))
	} else {
		reader = strings.NewReader("")
	}
	req, err := http.NewRequest(method, e.server.URL+path, reader)
	if err != nil {
		e.t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		e.t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	var payload map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&payload)
	return resp.StatusCode, payload
}

func dataOf(payload map[string]any) map[string]any {
	if v, okv := payload["data"].(map[string]any); okv {
		return v
	}
	return nil
}

func listOf(payload map[string]any) []any {
	if v, okv := payload["data"].([]any); okv {
		return v
	}
	return nil
}

// sseStream opens the stream endpoint and collects events until [DONE].
func (e *env) sseStream(token, conversationID, content string) (status int, events []string) {
	e.t.Helper()
	url := fmt.Sprintf("%s/api/chat/conversations/%s/stream?token=%s&content=%s",
		e.server.URL, conversationID, token, strings.ReplaceAll(content, " ", "%20"))
	resp, err := http.Get(url)
	if err != nil {
		e.t.Fatalf("stream: %v", err)
	}
	defer resp.Body.Close()
	status = resp.StatusCode
	if status != http.StatusOK {
		return
	}
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "[DONE]" {
			events = append(events, "[DONE]")
			return
		}
		var obj map[string]any
		if err := json.Unmarshal([]byte(payload), &obj); err == nil {
			if delta, okv := obj["delta"].(string); okv {
				events = append(events, delta)
			} else if code, okv := obj["code"].(string); okv {
				events = append(events, "ERR:"+code)
				return
			}
		}
	}
	return
}
