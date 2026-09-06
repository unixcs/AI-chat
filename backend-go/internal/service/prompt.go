package service

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// maxPromptBytes caps a single prompt revision (≈ several thousand tokens).
const maxPromptBytes = 20000

// EffectiveSystemPrompt returns the prompt a NEW chat request should use:
// the highest promptRevisions row when the admin has ever saved one, else the
// startup configuration (env/file) value. Read per request — no cache, so a
// save is effective for the very next chat.
func (s *Service) EffectiveSystemPrompt() string {
	if rev, err := s.Store.CurrentPromptRevision(); err == nil {
		return rev.Content
	}
	return s.Cfg.SystemPrompt
}

// AdminPromptInfo is the GET /api/admin/prompt payload.
func (s *Service) AdminPromptInfo() map[string]any {
	rev, err := s.Store.CurrentPromptRevision()
	if err != nil {
		return map[string]any{
			"content": s.Cfg.SystemPrompt,
			"version": nil,
			"updatedAt": nil,
			"operator": nil,
			"source":   "env",
		}
	}
	return map[string]any{
		"content":   rev.Content,
		"version":   rev.Version,
		"updatedAt": rev.CreatedAt,
		"operator":  rev.Operator,
		"source":    "db",
	}
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
	version, err := s.Store.CreatePromptRevision(content, operator)
	if err != nil {
		return nil, fail(500, "服务器繁忙")
	}
	auditLog("update-prompt", fmt.Sprintf("version=%d bytes=%d", version, len(content)), operator)
	return s.AdminPromptInfo(), nil
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
func (s *Service) AdminRestorePrompt(version int64, operator string) (map[string]any, *ServiceError) {
	rev, err := s.Store.GetPromptRevision(version)
	if err != nil {
		return nil, fail(404, "该版本不存在")
	}
	newVersion, err := s.Store.CreatePromptRevision(rev.Content, operator)
	if err != nil {
		return nil, fail(500, "服务器繁忙")
	}
	auditLog("restore-prompt", fmt.Sprintf("from=%d to=%d bytes=%d", version, newVersion, len(rev.Content)), operator)
	return s.AdminPromptInfo(), nil
}

// BotPromptSet switches the effective prompt to a copy of the given version
// (Telegram-facing preset switch; same business rules as the Web restore).
func (s *Service) BotPromptSet(version int64, operator string) *ServiceError {
	_, serr := s.AdminRestorePrompt(version, operator)
	return serr
}
