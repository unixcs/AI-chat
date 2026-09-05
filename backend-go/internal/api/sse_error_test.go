package api

import (
	"testing"

	"ai-chat-backend/internal/ai"
)

// Table-driven lock of the Node-contract SSE error mapping (server.js:686-703).
func TestSseErrorFor(t *testing.T) {
	cases := []struct {
		code    string
		msg     string
		wantMsg string
	}{
		{"SESSION_KICKED", "任意", "账号已在其他设备登录"},
		{"MODEL_QUEUE_OVERFLOW", "x", "当前请求较多，请稍后再试"},
		{"MODEL_RATE_LIMIT", "x", "当前请求较多，请稍后再试"},
		{"MODEL_TIMEOUT", "x", "响应超时，请重试"},
		{"MODEL_INSUFFICIENT_BALANCE", "x", "服务额度不足，请联系管理员"},
		{"STREAM_ERROR", "模型响应中断", "模型响应中断"}, // default: pass-through message
		{"MODEL_UPSTREAM_ERROR", "", "模型调用失败"},  // default: fallback message
	}
	for _, c := range cases {
		code, msg := sseErrorFor(&ai.ModelError{Code: c.code, Message: c.msg})
		if code != c.code {
			t.Errorf("code %q: got %q", c.code, code)
		}
		if msg != c.wantMsg {
			t.Errorf("code %q: got message %q, want %q", c.code, msg, c.wantMsg)
		}
	}
}
