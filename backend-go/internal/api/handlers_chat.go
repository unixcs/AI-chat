package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"ai-chat-backend/internal/ai"
	"ai-chat-backend/internal/model"
	"ai-chat-backend/internal/service"
	"ai-chat-backend/internal/store"
)

func serviceMembershipValid(u *model.User) bool { return service.MembershipValid(u) }

// ---------- conversations ----------

func (a *API) handleListConversations(user *model.User, w http.ResponseWriter, r *http.Request) {
	page := queryInt(r, "page", 1)
	pageSize := queryInt(r, "pageSize", 100)
	list, serr := a.Svc.ListConversations(user.ID, page, pageSize)
	if serr != nil {
		fail(w, serr)
		return
	}
	ok(w, list)
}

func (a *API) handleCreateConversation(user *model.User, w http.ResponseWriter, r *http.Request) {
	c, serr := a.Svc.CreateConversation(user.ID)
	if serr != nil {
		fail(w, serr)
		return
	}
	ok(w, c)
}

func (a *API) handleDeleteConversation(user *model.User, w http.ResponseWriter, r *http.Request) {
	if serr := a.Svc.DeleteConversation(user.ID, r.PathValue("id")); serr != nil {
		fail(w, serr)
		return
	}
	ok(w, true)
}

func (a *API) handleGetMessages(user *model.User, w http.ResponseWriter, r *http.Request) {
	msgs, serr := a.Svc.GetMessages(user.ID, r.PathValue("id"))
	if serr != nil {
		fail(w, serr)
		return
	}
	ok(w, msgs)
}

// handlePostMessage keeps the legacy non-stream endpoint (demo reply) for
// contract parity with the Node backend.
func (a *API) handlePostMessage(user *model.User, w http.ResponseWriter, r *http.Request) {
	var body struct {
		Content string `json:"content"`
	}
	decodeJSON(w, r, &body)
	if body.Content == "" {
		failMsg(w, 400, "消息内容不能为空")
		return
	}
	conversationID := r.PathValue("id")
	c, err := a.Svc.Store.GetConversation(conversationID)
	if err != nil || c.UserID != user.ID {
		failMsg(w, 404, "会话不存在")
		return
	}
	if !membershipValid(user) {
		failMsg(w, 403, "会员过期，请续费后使用")
		return
	}
	now := store.Now()
	_ = a.Svc.Store.InsertMessage(&model.Message{
		ID: store.NewID(), ConversationID: c.ID, Role: "user", Content: body.Content, CreatedAt: now,
	})
	reply := &model.Message{
		ID: store.NewID(), ConversationID: c.ID, Role: "assistant",
		Content: "已收到你的问题：" + body.Content + "\n这是演示回复，后续可替换真实模型 API。",
		CreatedAt: store.Now(),
	}
	_ = a.Svc.Store.InsertMessage(reply)
	_ = a.Svc.Store.TouchConversation(c.ID)
	ok(w, reply)
}

// ---------- SSE stream ----------

// handleStream streams an AI reply as Server-Sent Events. Auth happens via
// Bearer header or token query param (EventSource cannot set headers).
func (a *API) handleStream(w http.ResponseWriter, r *http.Request) {
	token := bearerOrQuery(r)
	content := trimQuery(r.URL.Query().Get("content"))
	conversationID := r.PathValue("id")

	user, conv, messages, serr := a.Svc.PrepareStream(r.Context(), token, conversationID, content)
	if serr != nil {
		fail(w, serr)
		return
	}

	flusher, writable := sseHeaders(w)
	if !writable {
		return
	}
	_ = user
	_ = conv

	var mu sync.Mutex
	var assistantText string
	done := make(chan struct{})

	// The stream goroutine owns ALL SSE writes; the handler returns only after
	// it finishes (client cancel propagates through ctx and the router exits
	// promptly), so writes never race the http server tearing down the response.
	go func() {
		defer close(done)
		_, merr := a.Router.Stream(r.Context(), ai.ChatRequest{Messages: messages}, func(delta string) {
			mu.Lock()
			assistantText += delta
			mu.Unlock()
			sseWrite(flusher, w, mustJSON(map[string]any{"delta": delta}))
		}, nil)

		if merr != nil && r.Context().Err() == nil {
			code, message := sseErrorFor(merr)
			sseWrite(flusher, w, mustJSON(map[string]any{"code": code, "error": message}))
		}
	}()

	<-done

	// Persist semantics (design D4.5): partial content is always saved; the
	// placeholder is ONLY for client-initiated aborts with zero content. An
	// upstream failure with a healthy client persists nothing.
	mu.Lock()
	text := assistantText
	mu.Unlock()
	if text != "" {
		a.Svc.PersistReply(conversationID, text)
	} else if r.Context().Err() != nil {
		a.Svc.PersistReply(conversationID, "模型输出已中断。")
	}

	if r.Context().Err() != nil {
		return
	}
	sseRaw(flusher, w, "data: [DONE]\n\n")
}

// sseErrorFor mirrors the Node backend's SSE error mapping verbatim
// (backend/server.js catch block) so the frontend behaves identically.
func sseErrorFor(merr *ai.ModelError) (string, string) {
	switch merr.Code {
	case "SESSION_KICKED":
		return merr.Code, "账号已在其他设备登录"
	case "MODEL_QUEUE_OVERFLOW", "MODEL_RATE_LIMIT":
		return merr.Code, "当前请求较多，请稍后再试"
	case "MODEL_TIMEOUT":
		return merr.Code, "响应超时，请重试"
	case "MODEL_INSUFFICIENT_BALANCE":
		return merr.Code, "服务额度不足，请联系管理员"
	default:
		// Node: errorCode = error.modelCode || 'STREAM_ERROR'
		code := merr.Code
		if code == "" {
			code = "STREAM_ERROR"
		}
		if merr.Message != "" {
			return code, merr.Message
		}
		return code, "模型调用失败"
	}
}

func mustJSON(v any) string {
	var sb strings.Builder
	enc := json.NewEncoder(&sb)
	enc.SetEscapeHTML(false) // Node JSON.stringify does not escape <>& by default
	_ = enc.Encode(v)
	return strings.TrimRight(sb.String(), "\n")
}

func membershipValid(u *model.User) bool { return serviceMembershipValid(u) }

func bearerOrQuery(r *http.Request) string {
	token := bearerToken(r.Header.Get("Authorization"))
	if token == "" {
		token = r.URL.Query().Get("token")
	}
	return token
}

func bearerToken(header string) string {
	const prefix = "Bearer "
	if len(header) < len(prefix) || header[:len(prefix)] != prefix {
		return ""
	}
	return trimQuery(header[len(prefix):])
}

func trimQuery(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
