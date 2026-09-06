package store

import "database/sql"

// Preference prompt cards: per-option system-prompt fragments the admin can
// edit in the web console. Content lives in the settings table (one key per
// card); the presets below are seeded at startup with INSERT OR IGNORE, so a
// fresh database ships ready-to-use cards and an admin's edits survive
// restarts. The card list itself is code-owned (whitelist) — dimensions and
// values are never taken from user input.

type PrefPromptCard struct {
	Dimension  string `json:"dimension"`
	Value      string `json:"value"`
	Label      string `json:"label"`
	Content    string `json:"content"`    // effective content (settings row or preset)
	Preset     string `json:"preset"`     // factory default, shown for restore
	Customized bool   `json:"customized"` // true when a settings row exists
}

type prefPromptSpec struct {
	Dimension string
	Value     string
	Label     string
	Preset    string
}

// prefPromptSpecs is the single source of truth for card layout order.
var prefPromptSpecs = []prefPromptSpec{
	{"answerLength", "concise", "精简", "回答务必精简：控制在 100 字以内，直接给结论，不展开解释，不重复问题。"},
	{"answerLength", "standard", "适中", "回答篇幅适中：一般 100–300 字，先给结论再简要说明，用户追问时可以适当展开。"},
	{"answerLength", "detailed", "详细", "回答详细展开：分点说明，给出充分的解释、细节和示例，把问题讲透。"},
	{"answerStyle", "plain", "大白话", "全程用通俗易懂的大白话回答，避免专业术语；必须使用术语时，紧跟一句生活化的比喻解释。"},
	{"answerStyle", "standard", "标准", "用自然、清晰、友好的语气回答，表达得体即可，不刻意口语化也不刻意书面化。"},
	{"answerStyle", "professional", "专业", "使用专业、准确的表达方式回答：术语规范、逻辑清晰，可使用领域惯用说法，但避免堆砌行话。"},
	{"answerStyle", "rigorous", "严谨", "回答保持严谨：区分事实与推测，不确定的内容明确说明；给出结论时附上理由或依据，不臆断。"},
	{"answerStyle", "encouraging", "鼓励", "语气温和、以鼓励为主：肯定用户的思路与努力，指出改进方向时先扬后抑，给出具体可行的建议。"},
	{"answerFormat", "standard", "标准", "排版方式自由：根据内容选用最合适的排版（分段、列表、表格、代码块等），保证阅读体验。"},
	{"answerFormat", "plain", "纯文字", "只输出纯文字：不使用 Markdown、列表、表格、标题、代码块、加粗、斜体、emoji 等任何特殊排版符号，用自然段落书写，保持纯文本阅读体验。"},
}

// ValidPrefCard reports whether dimension/value form a known card.
func ValidPrefCard(dimension, value string) bool {
	for _, sp := range prefPromptSpecs {
		if sp.Dimension == dimension && sp.Value == value {
			return true
		}
	}
	return false
}

func prefPromptKey(dimension, value string) string {
	return "prefPrompt." + dimension + "." + value
}

// SeedPrefPrompts inserts the factory defaults without overwriting admin edits.
func (s *Store) SeedPrefPrompts() error {
	for _, sp := range prefPromptSpecs {
		if _, err := s.DB.Exec(
			`INSERT OR IGNORE INTO settings (key, value) VALUES (?, ?)`,
			prefPromptKey(sp.Dimension, sp.Value), sp.Preset,
		); err != nil {
			return err
		}
	}
	return nil
}

// ListPrefPrompts returns every card in spec order with effective content.
// A card counts as customized when its stored value differs from the factory
// preset (seeding stores presets as rows, so row-existence is not enough).
func (s *Store) ListPrefPrompts() ([]PrefPromptCard, error) {
	out := make([]PrefPromptCard, 0, len(prefPromptSpecs))
	for _, sp := range prefPromptSpecs {
		var stored sql.NullString
		probeErr := s.DB.QueryRow(`SELECT value FROM settings WHERE key = ?`, prefPromptKey(sp.Dimension, sp.Value)).Scan(&stored)
		content := sp.Preset
		customized := false
		if probeErr == nil {
			content = stored.String
			customized = stored.String != sp.Preset
		}
		out = append(out, PrefPromptCard{
			Dimension:  sp.Dimension,
			Value:      sp.Value,
			Label:      sp.Label,
			Content:    content,
			Preset:     sp.Preset,
			Customized: customized,
		})
	}
	return out, nil
}

// GetPrefPromptContent returns the effective content of one card.
func (s *Store) GetPrefPromptContent(dimension, value string) (string, error) {
	for _, sp := range prefPromptSpecs {
		if sp.Dimension == dimension && sp.Value == value {
			return s.GetSetting(prefPromptKey(sp.Dimension, sp.Value), sp.Preset), nil
		}
	}
	return "", ErrNotFound
}

// SetPrefPromptContent saves an admin-edited card.
func (s *Store) SetPrefPromptContent(dimension, value, content string) error {
	if !ValidPrefCard(dimension, value) {
		return ErrNotFound
	}
	return s.SetSetting(prefPromptKey(dimension, value), content)
}

// ResetPrefPromptContent removes the override so the card falls back to its
// factory preset.
func (s *Store) ResetPrefPromptContent(dimension, value string) error {
	if !ValidPrefCard(dimension, value) {
		return ErrNotFound
	}
	_, err := s.DB.Exec(`DELETE FROM settings WHERE key = ?`, prefPromptKey(dimension, value))
	return err
}
