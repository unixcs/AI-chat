// Package service holds all business logic. Both the HTTP API and the
// Telegram bot go through this layer — the bot never touches SQL directly.
package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"ai-chat-backend/internal/ai"
	"ai-chat-backend/internal/auth"
	"ai-chat-backend/internal/config"
	"ai-chat-backend/internal/model"
	"ai-chat-backend/internal/store"
)

type ServiceError struct {
	Status  int
	Code    string // optional machine code, mirrors Node (SESSION_KICKED, ...)
	Message string
}

func (e *ServiceError) Error() string { return e.Message }

func fail(status int, format string, args ...any) *ServiceError {
	return &ServiceError{Status: status, Message: fmt.Sprintf(format, args...)}
}

var ErrNoAnnouncement = store.ErrNotFound

var phoneRe = regexp.MustCompile(`^1\d{10}$`)

// AIStreamer is what service needs from the ai package; fakes can implement it.
type AIStreamer interface {
	Stream(ctx context.Context, req ai.ChatRequest, onDelta func(string), onModel func(string)) (*ai.StreamResult, *ai.ModelError)
	HasAPIKey() bool
	Mode() string
	HealthSnapshot() []ai.HealthSnapshot
	StatsSnapshot() map[string]int64
}

type Service struct {
	Cfg    *config.Config
	Store  *store.Store
	Router AIStreamer
}

func New(cfg *config.Config, st *store.Store, router AIStreamer) *Service {
	return &Service{Cfg: cfg, Store: st, Router: router}
}

// ---------- session helpers ----------

func (s *Service) validateUserSession(c *auth.Claims) (*model.User, *ServiceError) {
	user, err := s.Store.GetUserByID(c.UserID)
	if err != nil || user.Role != "user" {
		return nil, &ServiceError{Status: 401, Code: "USER_INVALID", Message: "用户不存在或已失效"}
	}
	if user.Status != "active" {
		return nil, &ServiceError{Status: 403, Code: "USER_DISABLED", Message: "账号已被禁用"}
	}
	if c.SessionID == "" || user.CurrentSession == "" || c.SessionID != user.CurrentSession {
		return nil, &ServiceError{Status: 401, Code: "SESSION_KICKED", Message: "账号已在其他设备登录"}
	}
	return user, nil
}

func (s *Service) AuthUser(bearer string) (*model.User, *ServiceError) {
	token := auth.ParseBearerToken(bearer)
	if token == "" {
		return nil, &ServiceError{Status: 401, Message: "未登录或登录已过期"}
	}
	claims, err := auth.VerifyToken(s.Cfg.JWTSecret, token)
	if err != nil {
		return nil, &ServiceError{Status: 401, Message: "用户鉴权失败"}
	}
	if claims.Type != "user" {
		return nil, &ServiceError{Status: 403, Message: "无效用户令牌"}
	}
	user, serr := s.validateUserSession(claims)
	if serr != nil {
		return nil, serr
	}
	return user, nil
}

// AuthUserForStream mirrors AuthUser but with the stream endpoint's error codes.
func (s *Service) AuthUserForStream(token string) (*model.User, *ServiceError) {
	if token == "" {
		return nil, &ServiceError{Status: 401, Message: "未登录或登录已过期"}
	}
	claims, err := auth.VerifyToken(s.Cfg.JWTSecret, token)
	if err != nil {
		return nil, &ServiceError{Status: 401, Code: "USER_AUTH_FAILED", Message: "用户鉴权失败"}
	}
	if claims.Type != "user" {
		return nil, &ServiceError{Status: 403, Code: "INVALID_USER_TOKEN", Message: "无效用户令牌"}
	}
	user, serr := s.validateUserSession(claims)
	if serr != nil {
		return nil, serr
	}
	return user, nil
}

func (s *Service) AuthAdmin(bearer string) (*model.User, *ServiceError) {
	token := auth.ParseBearerToken(bearer)
	if token == "" {
		return nil, &ServiceError{Status: 401, Message: "管理员未登录"}
	}
	claims, err := auth.VerifyToken(s.Cfg.JWTSecret, token)
	if err != nil {
		return nil, &ServiceError{Status: 401, Message: "管理员鉴权失败"}
	}
	if claims.Type != "admin" {
		return nil, &ServiceError{Status: 403, Message: "无效管理员令牌"}
	}
	admin, err := s.Store.GetUserByID(claims.AdminID)
	if err != nil || admin.Role != "admin" {
		return nil, &ServiceError{Status: 401, Message: "管理员鉴权失败"}
	}
	return admin, nil
}

// ---------- auth ----------

func (s *Service) Register(phone, nickname, password string) *ServiceError {
	if !phoneRe.MatchString(phone) {
		return fail(400, "手机号格式不正确")
	}
	if len(password) > 72 {
		return fail(400, "密码过长")
	}
	if strings.TrimSpace(nickname) == "" {
		return fail(400, "昵称不能为空")
	}
	if utf8.RuneCountInString(password) < 6 {
		return fail(400, "密码至少 6 位")
	}
	if !s.RegistrationOpen() {
		return fail(403, "当前已关闭注册")
	}
	hash, err := store.HashPassword(password)
	if err != nil {
		return fail(500, "服务器繁忙")
	}
	now := store.Now()
	u := &model.User{
		ID:           store.NewID(),
		Phone:        phone,
		Nickname:     nickname,
		PasswordHash: hash,
		Status:       "active",
		Role:         "user",
		CreatedAt:    now,
		LastLoginAt:  now,
	}
	// Insert-first: the UNIQUE index on users.phone makes the duplicate check
	// atomic, so concurrent registrations cannot create twin accounts.
	if err := s.Store.CreateUser(u); err != nil {
		if store.IsUniqueViolation(err) {
			return fail(400, "手机号已注册")
		}
		return fail(500, "服务器繁忙")
	}
	return nil
}

func (s *Service) Login(phone, password string) (string, *ServiceError) {
	user, err := s.Store.GetUserByPhone(phone)
	if err != nil {
		return "", fail(400, "账号不存在")
	}
	if user.Role != "user" {
		return "", fail(400, "账号不存在")
	}
	if user.Status != "active" {
		return "", fail(403, "账号已被禁用")
	}
	if !store.CheckPassword(user.PasswordHash, password) {
		return "", fail(400, "密码错误")
	}
	sessionID := newSessionID()
	if serr := s.Store.UpdateUserLogin(user.ID, sessionID); serr != nil {
		return "", fail(500, "服务器繁忙")
	}
	token, err := auth.SignToken(s.Cfg.JWTSecret, auth.Claims{UserID: user.ID, SessionID: sessionID, Type: "user"})
	if err != nil {
		return "", fail(500, "服务器繁忙")
	}
	return token, nil
}

func (s *Service) Logout(userID string) *ServiceError {
	if err := s.Store.UpdateUserSession(userID, ""); err != nil {
		return fail(500, "服务器繁忙")
	}
	return nil
}

func (s *Service) UpdateProfile(userID, nickname, avatarURL string) (*model.User, *ServiceError) {
	user, err := s.Store.GetUserByID(userID)
	if err != nil {
		return nil, fail(404, "用户不存在")
	}
	nextNickname := user.Nickname
	if strings.TrimSpace(nickname) != "" {
		nextNickname = nickname
	}
	if err := s.Store.UpdateUserProfile(userID, nextNickname, avatarURL); err != nil {
		return nil, fail(500, "服务器繁忙")
	}
	updated, err := s.Store.GetUserByID(userID)
	if err != nil {
		return nil, fail(500, "服务器繁忙")
	}
	return updated, nil
}

func (s *Service) ChangePassword(userID, oldPassword, newPassword string) *ServiceError {
	user, err := s.Store.GetUserByID(userID)
	if err != nil {
		return fail(404, "用户不存在")
	}
	if len(newPassword) > 72 {
		return fail(400, "密码过长")
	}
	if !store.CheckPassword(user.PasswordHash, oldPassword) {
		return fail(400, "旧密码错误")
	}
	hash, err := store.HashPassword(newPassword)
	if err != nil {
		return fail(500, "服务器繁忙")
	}
	if err := s.Store.UpdateUserPassword(userID, hash); err != nil {
		return fail(500, "服务器繁忙")
	}
	return nil
}

func (s *Service) UpdatePreferences(userID, length, style string) *ServiceError {
	if length != "" && !validAnswerLength(length) {
		return fail(400, "回答长度参数错误")
	}
	if style != "" && !validAnswerStyle(style) {
		return fail(400, "回答风格参数错误")
	}
	if err := s.Store.UpdateUserPreferences(userID, length, style); err != nil {
		return fail(500, "服务器繁忙")
	}
	return nil
}

func validAnswerLength(v string) bool {
	switch v {
	case "concise", "standard", "detailed":
		return true
	}
	return false
}

func validAnswerStyle(v string) bool {
	switch v {
	case "plain", "standard", "professional", "rigorous", "encouraging":
		return true
	}
	return false
}

// ---------- redeem ----------

func (s *Service) Redeem(userID, code string) (*model.User, *ServiceError) {
	if !s.RedeemOpen() {
		return nil, fail(403, "当前已关闭兑换码兑换")
	}
	user, err := s.Store.GetUserByID(userID)
	if err != nil {
		return nil, fail(404, "用户不存在")
	}
	redeemCode, err := s.Store.GetRedeemCodeByCode(strings.TrimSpace(code))
	if err != nil {
		return nil, fail(400, "兑换码不存在")
	}
	if redeemCode.Status != "unused" {
		return nil, fail(400, "兑换码已失效或已使用")
	}

	if serr := s.Store.Redeem(redeemCode.ID, redeemCode.Code, userID, user.Phone, redeemCode.DurationMonths); serr != nil {
		return nil, fail(400, "兑换码已失效或已使用")
	}
	updated, err := s.Store.GetUserByID(userID)
	if err != nil {
		return nil, fail(500, "服务器繁忙")
	}
	return updated, nil
}

// MembershipValid reports whether the user's membership is active.
func MembershipValid(u *model.User) bool {
	if u.MemberExpireAt == "" {
		return false
	}
	return store.TimeValue(u.MemberExpireAt).After(time.Now())
}

// ---------- settings ----------

func (s *Service) RegistrationOpen() bool {
	v := s.Store.GetSetting("registrationOpen", "")
	if v == "" {
		return s.Cfg.RegistrationOpen
	}
	return v == "true"
}

func (s *Service) RedeemOpen() bool {
	v := s.Store.GetSetting("redeemOpen", "")
	if v == "" {
		return s.Cfg.RedeemOpen
	}
	return v == "true"
}

func (s *Service) SetRegistrationOpen(open bool, operator string) error {
	auditLog("toggle-registration", fmt.Sprintf("%t", open), operator)
	return s.Store.SetSetting("registrationOpen", fmt.Sprintf("%t", open))
}

func (s *Service) SetRedeemOpen(open bool, operator string) error {
	auditLog("toggle-redeem", fmt.Sprintf("%t", open), operator)
	return s.Store.SetSetting("redeemOpen", fmt.Sprintf("%t", open))
}

// auditLog is the minimal audit trail for sensitive admin operations
// (original Plan §19: 重要管理操作记录日志). Both the HTTP admin API ("web-admin")
// and the Telegram bot ("tg:<id>") funnel their mutations through service
// methods, so every audit line carries its true operator identity.
func auditLog(action, detail, operator string) {
	by := operator
	if by == "" {
		by = "web-admin"
	}
	fmt.Printf("[audit] operator=%s action=%s detail=%s at=%s\n", by, action, detail, store.Now())
}

// ---------- misc ----------

func newSessionID() string {
	return fmt.Sprintf("%d_%s", time.Now().UnixMilli(), randLower(10))
}

func randLower(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		b[i] = chars[idx.Int64()]
	}
	return string(b)
}

func randUpper(n int) string {
	return strings.ToUpper(randLower(n))
}

func adminToken(secret, adminID string) (string, *ServiceError) {
	token, err := auth.SignToken(secret, auth.Claims{AdminID: adminID, Type: "admin"})
	if err != nil {
		return "", fail(500, "服务器繁忙")
	}
	return token, nil
}
