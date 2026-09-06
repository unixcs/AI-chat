package service

import (
	"context"
	"sort"
	"strings"
	"time"

	"ai-chat-backend/internal/ai"
	"ai-chat-backend/internal/model"
	"ai-chat-backend/internal/store"
)

// ---------- conversations ----------

// ListConversations reproduces the Node pagination contract verbatim:
// page: max(1, parseInt||1); pageSize: min(500, parseInt||100) — including
// the quirky but real behavior that a negative pageSize flows into SQLite's
// LIMIT (negative = unlimited) instead of being clamped.
func (s *Service) ListConversations(userID string, page, pageSize int) ([]model.Conversation, *ServiceError) {
	if page < 1 {
		page = 1
	}
	if page > maxPage {
		page = maxPage
	}
	switch {
	case pageSize > 500:
		pageSize = 500
	case pageSize == 0:
		pageSize = 100
	}
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0 // SQLite treats negative OFFSET as 0
	}
	list, err := s.Store.ListConversationsByUser(userID, pageSize, offset)
	if err != nil {
		return nil, fail(500, "服务器繁忙")
	}
	return list, nil
}

func (s *Service) CreateConversation(userID string) (*model.Conversation, *ServiceError) {
	now := store.Now()
	c := &model.Conversation{
		ID:        store.NewID(),
		UserID:    userID,
		Title:     "新对话 " + time.Now().UTC().Format("01-02 15:04"),
		CreatedAt: now,
		UpdatedAt: now,
	}
	// Node formatted the title with the server-local timezone (UTC on the
	// containers); +8 makes the title readable for the actual user base.
	c.Title = "新对话 " + time.Now().In(time.FixedZone("CST", 8*3600)).Format("01-02 15:04")
	if err := s.Store.CreateConversation(c); err != nil {
		return nil, fail(500, "服务器繁忙")
	}
	return c, nil
}

// DeleteConversation removes a conversation owned by the user.
func (s *Service) DeleteConversation(userID, conversationID string) *ServiceError {
	c, err := s.Store.GetConversation(conversationID)
	if err != nil || c.UserID != userID {
		return fail(404, "会话不存在")
	}
	if err := s.Store.DeleteConversation(conversationID); err != nil {
		return fail(500, "服务器繁忙")
	}
	return nil
}

func (s *Service) GetMessages(userID, conversationID string) ([]model.Message, *ServiceError) {
	c, err := s.Store.GetConversation(conversationID)
	if err != nil || c.UserID != userID {
		return nil, fail(404, "会话不存在")
	}
	msgs, err := s.Store.GetMessages(conversationID)
	if err != nil {
		return nil, fail(500, "服务器繁忙")
	}
	return msgs, nil
}

// PrepareStream validates a stream request and inserts the user message.
// It returns everything the caller needs to run the model and persist the reply.
func (s *Service) PrepareStream(ctx context.Context, token, conversationID, content string) (*model.User, *model.Conversation, []ai.ChatMessage, *ServiceError) {
	user, serr := s.AuthUserForStream(token)
	if serr != nil {
		return nil, nil, nil, serr
	}

	content = strings.TrimSpace(content)
	if content == "" {
		return nil, nil, nil, fail(400, "消息内容不能为空")
	}

	c, err := s.Store.GetConversation(conversationID)
	if err != nil || c.UserID != user.ID {
		return nil, nil, nil, fail(404, "会话不存在")
	}
	if !MembershipValid(user) {
		return nil, nil, nil, fail(403, "会员过期，请续费后使用")
	}
	if !s.Router.HasAPIKey() {
		return nil, nil, nil, fail(500, "服务器未配置 DEEPSEEK_API_KEY")
	}

	// History snapshot BEFORE the new user message lands (matches Node order).
	history, err := s.Store.GetMessages(conversationID)
	if err != nil {
		return nil, nil, nil, fail(500, "服务器繁忙")
	}
	contextMessages := buildContextByRounds(history, 6)

	if serr := s.insertUserMessage(c, content); serr != nil {
		return nil, nil, nil, serr
	}

	messages := make([]ai.ChatMessage, 0, len(contextMessages)+2)
	messages = append(messages, ai.ChatMessage{Role: "system", Content: s.BuildSystemPrompt(user)})
	messages = append(messages, contextMessages...)
	messages = append(messages, ai.ChatMessage{Role: "user", Content: content})
	return user, c, messages, nil
}

func (s *Service) insertUserMessage(c *model.Conversation, content string) *ServiceError {
	msg := &model.Message{
		ID:             store.NewID(),
		ConversationID: c.ID,
		Role:           "user",
		Content:        content,
		CreatedAt:      store.Now(),
	}
	if err := s.Store.InsertMessage(msg); err != nil {
		return fail(500, "服务器繁忙")
	}
	if err := s.Store.TouchConversation(c.ID); err != nil {
		return fail(500, "服务器繁忙")
	}
	return nil
}

// PersistReply stores the assistant reply (or the partial text on abort) and
// touches the conversation. If the conversation was deleted while the stream
// was in flight, the reply is dropped instead of becoming an orphan row.
// bestEffort: DB failure must not break the stream, but it IS logged.
func (s *Service) PersistReply(conversationID, content string) {
	if content == "" {
		content = "模型输出已中断。"
	}
	if _, err := s.Store.GetConversation(conversationID); err != nil {
		return // deleted mid-stream
	}
	if err := s.Store.InsertMessage(&model.Message{
		ID:             store.NewID(),
		ConversationID: conversationID,
		Role:           "assistant",
		Content:        content,
		CreatedAt:      store.Now(),
	}); err != nil {
		println("[persist] assistant reply insert failed:", err.Error())
	}
	_ = s.Store.TouchConversation(conversationID)
}

// BuildSystemPrompt layers the admin-managed base prompt (hot-reloaded from
// SQLite; falls back to the startup config) with the user's three answer
// preference dimensions (length / style / format), each resolved from an
// admin-editable prompt card. Called once per chat request — a stream keeps
// the prompt it started with. Preferences live at the session layer; the
// user's own message is never edited.
func (s *Service) BuildSystemPrompt(user *model.User) string {
	prompt := s.EffectiveSystemPrompt()
	if prompt == "" {
		prompt = s.Cfg.SystemPromptDefault
	}
	prompt = s.appendPrefPrompt(prompt, "回答长度要求", "answerLength", user.AnswerLength)
	prompt = s.appendPrefPrompt(prompt, "回答风格要求", "answerStyle", user.AnswerStyle)
	prompt = s.appendPrefPrompt(prompt, "输出格式要求", "answerFormat", user.AnswerFormat)
	return prompt
}

// appendPrefPrompt appends one labeled preference segment. A missing or empty
// card skips the segment entirely; a transient read failure is logged and
// skipped rather than failing the chat.
func (s *Service) appendPrefPrompt(prompt, label, dimension, value string) string {
	content := s.GetPrefPromptContent(dimension, value)
	if content == "" {
		return prompt
	}
	return prompt + "\n\n" + label + "：" + content
}

// buildContextByRounds mirrors the Node backend: pair user/assistant messages
// into rounds, keep the last N rounds, drop dangling trailing assistant pairs.
func buildContextByRounds(history []model.Message, roundCount int) []ai.ChatMessage {
	sort.Slice(history, func(i, j int) bool { return history[i].CreatedAt < history[j].CreatedAt })

	type pair []model.Message
	var rounds []pair
	var pending *model.Message

	for i := range history {
		msg := history[i]
		if msg.Role != "user" && msg.Role != "assistant" {
			continue
		}
		if msg.Role == "user" {
			if pending != nil {
				rounds = append(rounds, pair{*pending})
			}
			m := msg
			pending = &m
			continue
		}
		if pending != nil {
			rounds = append(rounds, pair{*pending, msg})
			pending = nil
		}
	}
	if pending != nil {
		rounds = append(rounds, pair{*pending})
	}

	if len(rounds) > roundCount {
		rounds = rounds[len(rounds)-roundCount:]
	}

	var out []ai.ChatMessage
	for _, r := range rounds {
		for _, m := range r {
			out = append(out, ai.ChatMessage{Role: m.Role, Content: m.Content})
		}
	}
	return out
}

// ---------- announcements (user side) ----------

func (s *Service) CurrentAnnouncement(userID string) (*model.Announcement, *ServiceError) {
	a, err := s.Store.CurrentUnreadAnnouncement(userID)
	if err != nil {
		return nil, fail(404, "暂无公告")
	}
	return a, nil
}

// AckAnnouncement marks an announcement read. asOf (optional) is the
// updatedAt the user actually saw: when the admin replaced the announcement
// content in the meantime the ack is dropped, leaving the new content unread.
func (s *Service) AckAnnouncement(userID, announcementID, asOf string) *ServiceError {
	if _, err := s.Store.GetAnnouncement(announcementID); err != nil {
		return fail(404, "公告不存在")
	}
	if err := s.Store.AckAnnouncement(announcementID, userID, asOf); err != nil {
		return fail(500, "服务器繁忙")
	}
	return nil
}
