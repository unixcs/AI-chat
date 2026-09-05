package store

import "ai-chat-backend/internal/model"

// test fixtures (store package tests)

func userFixture(id, phone, hash string) *model.User {
	return &model.User{
		ID: id, Phone: phone, Nickname: "u-" + id, PasswordHash: hash,
		Status: "active", Role: "user", CreatedAt: Now(), LastLoginAt: Now(),
	}
}

func redeemFixture(id, code string, months int) *model.RedeemCode {
	return &model.RedeemCode{
		ID: id, Code: code, DurationMonths: months, Status: "unused", CreatedAt: Now(),
	}
}

func convFixture(id, userID string) *model.Conversation {
	return &model.Conversation{ID: id, UserID: userID, Title: "t", CreatedAt: Now(), UpdatedAt: Now()}
}

func msgFixture(id, convID, role, content string) *model.Message {
	return &model.Message{ID: id, ConversationID: convID, Role: role, Content: content, CreatedAt: Now()}
}

func annFixture(id, title, content string, active bool) *model.Announcement {
	return &model.Announcement{ID: id, Title: title, Content: content, Active: active, CreatedAt: Now(), UpdatedAt: Now()}
}
