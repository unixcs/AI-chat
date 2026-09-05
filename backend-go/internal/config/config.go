package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ProviderEntry describes one upstream OpenAI-compatible endpoint.
// The business layer never sees these; only the AI router does.
type ProviderEntry struct {
	Name                string `json:"name"`
	BaseURL             string `json:"baseURL"`
	APIKey              string `json:"apiKey"`
	Model               string `json:"model"`
	FirstTokenTimeoutMs int    `json:"firstTokenTimeoutMs,omitempty"`
	ExtraBodyRaw        string `json:"-"`
	Fallback            bool   `json:"fallback,omitempty"`
}

type Config struct {
	Port                   string
	SQLITEPath             string
	JWTSecret              string
	IsProduction           bool
	SystemPrompt           string
	SystemPromptDefault    string
	ExtraBody              string // DEEPSEEK_EXTRA_BODY raw JSON, applied to official entries
	AIModel                string
	AIBaseURL              string
	AIAPIKey               string
	AIMode                 string // gateway | official
	AIProviders            []ProviderEntry
	AIFirstTokenTimeoutMs  int
	AIGatewayBudgetMs      int
	AITotalTimeoutMs       int
	ModelConcurrency       int
	ModelQueueMax          int
	RegistrationOpen       bool
	RedeemOpen             bool
	TGBotToken             string
	TGAdminIDs             []int64
	TGProxy                string
}

func envStr(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}

func envInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

func envBool(key string, fallback bool) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if v == "" {
		return fallback
	}
	return v == "true" || v == "1" || v == "yes"
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:                  envStr("PORT", "3001"),
		SQLITEPath:            envStr("SQLITE_PATH", "data.sqlite"),
		JWTSecret:             envStr("JWT_SECRET", "demo_secret_key_2026"),
		IsProduction:          envStr("NODE_ENV", "") == "production",
		AIModel:               envStr("DEEPSEEK_MODEL", "deepseek-v4-flash"),
		AIBaseURL:             envStr("DEEPSEEK_BASE_URL", "https://api.deepseek.com"),
		AIAPIKey:              envStr("DEEPSEEK_API_KEY", ""),
		AIMode:                strings.ToLower(envStr("AI_MODE", "official")),
		ExtraBody:             strings.TrimSpace(os.Getenv("DEEPSEEK_EXTRA_BODY")),
		AIFirstTokenTimeoutMs: envInt("AI_FIRST_TOKEN_TIMEOUT_MS", 2000),
		AIGatewayBudgetMs:     envInt("AI_GATEWAY_BUDGET_MS", 6000),
		AITotalTimeoutMs:      envInt("DEEPSEEK_TIMEOUT_MS", 90000),
		ModelConcurrency:      envInt("MODEL_CONCURRENCY", 6),
		ModelQueueMax:         envInt("MODEL_QUEUE_MAX", 50),
		RegistrationOpen:      envBool("REGISTRATION_OPEN", true),
		RedeemOpen:            envBool("REDEEM_OPEN", true),
		TGBotToken:            envStr("TG_BOT_TOKEN", ""),
		TGProxy:               envStr("TG_PROXY", ""),
	}

	if cfg.AIMode != "gateway" && cfg.AIMode != "official" {
		return nil, fmt.Errorf("AI_MODE must be gateway or official, got %q", cfg.AIMode)
	}

	if !cfg.IsProduction {
		if abs, err := filepath.Abs(cfg.SQLITEPath); err == nil {
			cfg.SQLITEPath = abs
		}
	}
	if !filepath.IsAbs(cfg.SQLITEPath) && cfg.IsProduction {
		// Docker deployments always pass an absolute path; keep as-is.
		dir, err := os.Getwd()
		if err == nil {
			cfg.SQLITEPath = filepath.Join(dir, cfg.SQLITEPath)
		}
	}

	cfg.SystemPromptDefault = "你是一个专业、简洁、友好的中文 AI 助手。"
	cfg.SystemPrompt = loadSystemPrompt(envStr("DEEPSEEK_SYSTEM_PROMPT_FILE", ""), cfg.SystemPromptDefault)

	if cfg.JWTSecret == "demo_secret_key_2026" {
		fmt.Println("[auth] JWT_SECRET 未设置,正在使用内置默认密钥(不安全,生产环境必须在 .env 中配置)")
	}

	entries, err := loadProviders(cfg)
	if err != nil {
		return nil, err
	}
	cfg.AIProviders = entries

	if raw := envStr("TG_ADMIN_IDS", ""); raw != "" {
		for _, part := range strings.Split(raw, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			id, err := strconv.ParseInt(part, 10, 64)
			if err != nil {
				continue
			}
			cfg.TGAdminIDs = append(cfg.TGAdminIDs, id)
		}
	}

	return cfg, nil
}

func loadSystemPrompt(path, fallback string) string {
	if path == "" {
		return fallback
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(".", path)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return fallback
	}
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" {
		return fallback
	}
	return trimmed
}

// loadProviders builds the ordered model pool.
// Priority: AI_PROVIDERS_JSON (explicit pool) > single official entry from DEEPSEEK_* vars.
// An entry marked fallback:true is the official-API backstop; it always runs last
// and is the only entry in official mode.
func loadProviders(cfg *Config) ([]ProviderEntry, error) {
	var entries []ProviderEntry

	raw := strings.TrimSpace(os.Getenv("AI_PROVIDERS_JSON"))
	if raw != "" {
		if err := json.Unmarshal([]byte(raw), &entries); err != nil {
			return nil, fmt.Errorf("AI_PROVIDERS_JSON parse failed: %w", err)
		}
		for i := range entries {
			if entries[i].Name == "" || entries[i].BaseURL == "" || entries[i].Model == "" {
				return nil, fmt.Errorf("AI_PROVIDERS_JSON entry #%d missing name/baseURL/model", i+1)
			}
			entries[i].BaseURL = strings.TrimRight(entries[i].BaseURL, "/")
			if entries[i].FirstTokenTimeoutMs <= 0 {
				entries[i].FirstTokenTimeoutMs = cfg.AIFirstTokenTimeoutMs
			}
		}
	}

	// Official DeepSeek entry: always exists, built from DEEPSEEK_* envs.
	official := ProviderEntry{
		Name:                "deepseek-official",
		BaseURL:             strings.TrimRight(cfg.AIBaseURL, "/"),
		APIKey:              cfg.AIAPIKey,
		Model:               cfg.AIModel,
		FirstTokenTimeoutMs: cfg.AIFirstTokenTimeoutMs,
		ExtraBodyRaw:        cfg.ExtraBody,
		Fallback:            true,
	}

	// Drop any JSON-declared official duplicates, then append the canonical one last.
	filtered := entries[:0]
	for _, e := range entries {
		if e.Fallback {
			continue
		}
		filtered = append(filtered, e)
	}
	entries = append(filtered, official)

	if cfg.AIMode == "official" {
		return []ProviderEntry{official}, nil
	}
	return entries, nil
}

func (c *Config) ValidateRuntime() []string {
	var warnings []string
	if c.AIAPIKey == "" {
		warnings = append(warnings, "DEEPSEEK_API_KEY 未配置，聊天接口将返回 500")
	}
	return warnings
}
