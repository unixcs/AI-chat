package service

import (
	"fmt"
	"log"
	"strings"
	"unicode/utf8"

	"ai-chat-backend/internal/store"
)

// maxPrefPromptBytes caps a single preference prompt card.
const maxPrefPromptBytes = 4000

// AdminListPrefPrompts returns the 10 preference prompt cards in fixed order.
func (s *Service) AdminListPrefPrompts() ([]store.PrefPromptCard, *ServiceError) {
	cards, err := s.Store.ListPrefPrompts()
	if err != nil {
		return nil, fail(500, "服务器繁忙")
	}
	return cards, nil
}

// AdminUpdatePrefPrompt saves an admin-edited card; empty content is allowed
// and means "no extra instruction for this option".
func (s *Service) AdminUpdatePrefPrompt(dimension, value, content, operator string) *ServiceError {
	if !store.ValidPrefCard(dimension, value) {
		return fail(404, "该偏好项不存在")
	}
	content = strings.TrimSpace(content)
	if len(content) > maxPrefPromptBytes {
		return fail(400, "提示词过长，最多 %d 字节", maxPrefPromptBytes)
	}
	if !utf8.ValidString(content) {
		return fail(400, "提示词内容编码不正确")
	}
	if err := s.Store.SetPrefPromptContent(dimension, value, content); err != nil {
		return fail(500, "服务器繁忙")
	}
	auditLog("update-pref-prompt", fmt.Sprintf("%s.%s bytes=%d", dimension, value, len(content)), operator)
	return nil
}

// AdminResetPrefPrompt drops the override so the card returns to its factory
// preset.
func (s *Service) AdminResetPrefPrompt(dimension, value, operator string) *ServiceError {
	if !store.ValidPrefCard(dimension, value) {
		return fail(404, "该偏好项不存在")
	}
	if err := s.Store.ResetPrefPromptContent(dimension, value); err != nil {
		return fail(500, "服务器繁忙")
	}
	auditLog("reset-pref-prompt", dimension+"."+value, operator)
	return nil
}

// GetPrefPromptContent returns the effective content of one card ("" on
// unknown cards — callers skip empty segments anyway).
func (s *Service) GetPrefPromptContent(dimension, value string) string {
	if value == "" {
		value = "standard"
	}
	if !store.ValidPrefCard(dimension, value) {
		return ""
	}
	content, err := s.Store.GetPrefPromptContent(dimension, value)
	if err != nil {
		log.Printf("[pref-prompt] 读取卡片 %s.%s 失败，跳过该段: %v", dimension, value, err)
		return ""
	}
	return strings.TrimSpace(content)
}
