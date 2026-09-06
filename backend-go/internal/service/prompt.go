package service

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"unicode/utf8"

	"ai-chat-backend/internal/model"
	"ai-chat-backend/internal/store"
)

// maxPromptBytes caps a single prompt revision (≈ several thousand tokens).
const maxPromptBytes = 20000

// EffectiveSystemPrompt returns the prompt a NEW chat request should use:
// the highest promptRevisions row when the admin has ever saved one, else the
// startup configuration (env/file) value. Read per request — no cache, so a
// save is effective for the very next chat. A transient read failure falls
// back to the env value rather than failing chats (logged for observability).
func (s *Service) EffectiveSystemPrompt() string {
	rev, err := s.Store.CurrentPromptRevision()
	if err != nil {
		if !errors.Is(err, store.ErrNotFound) {
			log.Printf("[prompt] 读取当前提示词版本失败，本次回退环境配置: %v", err)
		}
		return s.Cfg.SystemPrompt
	}
	return rev.Content
}

// promptInfoFromRev is the {code:0,data} payload built from exactly the row a
// write created — never re-read the head, which under concurrency may already
// be someone else's version.
func promptInfoFromRev(rev *model.PromptRevision) map[string]any {
	return map[string]any{
		"content":   rev.Content,
		"version":   rev.Version,
		"updatedAt": rev.CreatedAt,
		"operator":  rev.Operator,
		"source":    "db",
	}
}

func envPromptInfo(cfgPrompt string) map[string]any {
	return map[string]any{
		"content":   cfgPrompt,
		"version":   nil,
		"updatedAt": nil,
		"operator":  nil,
		"source":    "env",
	}
}

// AdminPromptInfo is the GET /api/admin/prompt payload.
func (s *Service) AdminPromptInfo() map[string]any {
	rev, err := s.Store.CurrentPromptRevision()
	if err != nil {
		if !errors.Is(err, store.ErrNotFound) {
			log.Printf("[prompt] 读取当前提示词版本失败，回退环境配置展示: %v", err)
		}
		return envPromptInfo(s.Cfg.SystemPrompt)
	}
	return promptInfoFromRev(rev)
}

// AdminUpdatePrompt validates and stores a new prompt version.
func (s *Service) AdminUpdatePrompt(content, operator string) (map[string]any, *ServiceError) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, fail(400, "提示词内容不能为空")
	}
	if len(content) > maxPromptBytes {
		return nil, fail(400, "提示词过长，最多 %d 字节", maxPromptBytes)
	}
	if !utf8.ValidString(content) {
		return nil, fail(400, "提示词内容编码不正确")
	}
	rev, err := s.Store.CreatePromptRevision(content, operator)
	if err != nil {
		return nil, fail(500, "服务器繁忙")
	}
	auditLog("update-prompt", fmt.Sprintf("version=%d bytes=%d", rev.Version, len(content)), operator)
	return promptInfoFromRev(rev), nil
}

// AdminListPromptRevisions returns the newest 50 versions, newest first.
func (s *Service) AdminListPromptRevisions() ([]map[string]any, *ServiceError) {
	revs, err := s.Store.ListPromptRevisions(50)
	if err != nil {
		return nil, fail(500, "服务器繁忙")
	}
	out := make([]map[string]any, 0, len(revs))
	for i := range revs {
		r := &revs[i]
		out = append(out, map[string]any{
			"version":   r.Version,
			"content":   r.Content,
			"operator":  r.Operator,
			"createdAt": r.CreatedAt,
		})
	}
	return out, nil
}

// AdminRestorePrompt copies an old version's content as a NEW head version —
// history is append-only, so "restore" never rewrites or removes rows.
// Telegram-facing preset switching (/prompt set) goes through here too.
func (s *Service) AdminRestorePrompt(version int64, operator string) (map[string]any, *ServiceError) {
	rev, err := s.Store.GetPromptRevision(version)
	if err != nil {
		return nil, fail(404, "该版本不存在")
	}
	newRev, err := s.Store.CreatePromptRevision(rev.Content, operator)
	if err != nil {
		return nil, fail(500, "服务器繁忙")
	}
	auditLog("restore-prompt", fmt.Sprintf("from=%d to=%d bytes=%d", version, newRev.Version, len(rev.Content)), operator)
	return promptInfoFromRev(newRev), nil
}
