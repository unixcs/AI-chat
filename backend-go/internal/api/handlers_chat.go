package api

import (
	"encoding/json"
	"net/http"

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
	if err := decodeJSON(w, r, &body); err != nil || body.Content == "" {
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

	clientGone := r.Context().Done()
	streamErrCh := make(chan error, 1)

	var assistantText string
	done := make(chan struct{})

	go func() {
		defer close(done)
		result, merr := a.Router.Stream(r.Context(), ai.ChatRequest{Messages: messages}, func(delta string) {
			assistantText += delta
			sseWrite(flusher, w, mustJSON(map[string]any{"delta": delta}))
		}, nil)

		if merr != nil {
			if clientHasContent(assistantText) && r.Context().Err() != nil {
				// client cancelled: nothing more to send
				return
			}
			if r.Context().Err() == nil {
				sseWrite(flusher, w, mustJSON(map[string]any{"code": merr.Code, "error": merr.Message}))
			}
			streamErrCh <- merr
			return
		}
		_ = result
	}()

	select {
	case <-done:
	case <-clientGone:
	}

	// Persist whatever the model produced (partial on abort), like Node did.
	if assistantText != "" || r.Context().Err() == nil {
		a.Svc.PersistReply(conversationID, assistantText)
	}

	if r.Context().Err() != nil {
		return
	}
	sseRaw(flusher, w, "data: [DONE]\n\n")
}

func clientHasContent(s string) bool { return s != "" }

func mustJSON(v any) string {
	raw, _ := json.Marshal(v)
	return string(raw)
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
