// Package api wires the service layer to HTTP. Status codes, error message
// strings and the {code:0,data} envelope mirror the previous Node backend so
// the existing Vue frontend works unchanged.
package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ai-chat-backend/internal/ai"
	"ai-chat-backend/internal/model"
	"ai-chat-backend/internal/service"
)

type API struct {
	Svc    *service.Service
	Router *ai.Router
}

func New(svc *service.Service, router *ai.Router) *API {
	return &API{Svc: svc, Router: router}
}

// ---------- envelope helpers ----------

func ok(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{"code": 0, "data": data})
}

func fail(w http.ResponseWriter, serr *service.ServiceError) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	body := map[string]any{"message": serr.Message}
	if serr.Code != "" {
		body["code"] = serr.Code
	}
	w.WriteHeader(serr.Status)
	_ = json.NewEncoder(w).Encode(body)
}

func failMsg(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"message": message})
}

const maxBodyBytes = 100 << 10 // Node express.json() default limit

// decodeJSON mirrors the Node backend's tolerant behavior: an empty or
// unparseable body yields zero-value fields and the endpoint's own validation
// rejects them (Node lets "undefined" flow into the same checks). Only an
// oversized payload is a hard error (413, like express.json's limit). Returns
// false when the response has already been written.
func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	defer r.Body.Close()
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	if err := dec.Decode(target); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			_ = json.NewEncoder(w).Encode(map[string]any{"message": "请求体过大"})
			return false
		}
	}
	return true
}

func queryInt(r *http.Request, key string, fallback int) int {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return n
}

func queryString(r *http.Request, key string) string {
	return strings.TrimSpace(r.URL.Query().Get(key))
}

// ---------- middleware ----------

// userAuth authenticates a user session and hands the user row to the handler.
func (a *API) userAuth(next func(user *model.User, w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, serr := a.Svc.AuthUser(r.Header.Get("Authorization"))
		if serr != nil {
			fail(w, serr)
			return
		}
		next(user, w, r)
	}
}

// adminAuth authenticates an admin session and hands the admin row through.
func (a *API) adminAuth(next func(admin *model.User, w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		admin, serr := a.Svc.AuthAdmin(r.Header.Get("Authorization"))
		if serr != nil {
			fail(w, serr)
			return
		}
		next(admin, w, r)
	}
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		if r.URL.Path != "/api/health" {
			log.Printf("[http] %s %s %s", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
		}
	})
}

// ---------- health ----------

func (a *API) handleHealth(w http.ResponseWriter, r *http.Request) {
	// Node answered with the bare object (no {code,data} envelope) — keep it.
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "service": "backend", "time": modelNow()})
}

func modelNow() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05.000Z07:00")
}

// ---------- routing ----------

func (a *API) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", a.handleHealth)

	mux.HandleFunc("POST /api/auth/register", a.handleRegister)
	mux.HandleFunc("POST /api/auth/login", a.handleLogin)
	mux.HandleFunc("POST /api/auth/logout", a.userAuth(a.handleLogout))

	mux.HandleFunc("GET /api/user/session-check", a.userAuth(handleSessionCheck))
	mux.HandleFunc("GET /api/user/profile", a.userAuth(a.handleGetProfile))
	mux.HandleFunc("PUT /api/user/profile", a.userAuth(a.handleUpdateProfile))
	mux.HandleFunc("PUT /api/user/password", a.userAuth(a.handleChangePassword))
	mux.HandleFunc("PUT /api/user/preferences", a.userAuth(a.handlePreferences))
	mux.HandleFunc("POST /api/user/redeem", a.userAuth(a.handleRedeem))

	mux.HandleFunc("GET /api/chat/conversations", a.userAuth(a.handleListConversations))
	mux.HandleFunc("POST /api/chat/conversations", a.userAuth(a.handleCreateConversation))
	mux.HandleFunc("DELETE /api/chat/conversations/{id}", a.userAuth(a.handleDeleteConversation))
	mux.HandleFunc("GET /api/chat/conversations/{id}/messages", a.userAuth(a.handleGetMessages))
	mux.HandleFunc("POST /api/chat/conversations/{id}/messages", a.userAuth(a.handlePostMessage))
	mux.HandleFunc("GET /api/chat/conversations/{id}/stream", a.handleStream)

	mux.HandleFunc("GET /api/announcements/current", a.userAuth(a.handleCurrentAnnouncement))
	mux.HandleFunc("POST /api/announcements/{id}/ack", a.userAuth(a.handleAckAnnouncement))

	mux.HandleFunc("POST /api/admin/auth/login", a.handleAdminLogin)
	mux.HandleFunc("PUT /api/admin/auth/password", a.adminAuth(a.handleAdminChangePassword))
	mux.HandleFunc("GET /api/admin/dashboard", a.adminAuth(a.handleAdminDashboard))
	mux.HandleFunc("GET /api/admin/users", a.adminAuth(a.handleAdminUsers))
	mux.HandleFunc("PUT /api/admin/users/{id}", a.adminAuth(a.handleAdminUpdateUser))
	mux.HandleFunc("PUT /api/admin/users/{id}/status", a.adminAuth(a.handleAdminUserStatus))
	mux.HandleFunc("PUT /api/admin/users/{id}/reset-password", a.adminAuth(a.handleAdminResetPassword))
	mux.HandleFunc("PUT /api/admin/users/{id}/member-expire-at", a.adminAuth(a.handleAdminMemberExpire))
	mux.HandleFunc("GET /api/admin/roles", a.adminAuth(a.handleAdminRoles))
	mux.HandleFunc("GET /api/admin/menus", a.adminAuth(a.handleAdminMenus))
	mux.HandleFunc("GET /api/admin/members", a.adminAuth(a.handleAdminMembers))
	mux.HandleFunc("GET /api/admin/redeem-codes", a.adminAuth(a.handleAdminRedeemCodes))
	mux.HandleFunc("GET /api/admin/redeem-codes/export", a.adminAuth(a.handleAdminRedeemCodesExport))
	mux.HandleFunc("POST /api/admin/redeem-codes/batch", a.adminAuth(a.handleAdminRedeemCodesBatch))
	mux.HandleFunc("PUT /api/admin/redeem-codes/{id}/void", a.adminAuth(a.handleAdminRedeemCodeVoid))
	mux.HandleFunc("GET /api/admin/redeem-records", a.adminAuth(a.handleAdminRedeemRecords))
	mux.HandleFunc("GET /api/admin/conversations", a.adminAuth(a.handleAdminConversations))
	mux.HandleFunc("GET /api/admin/conversations/{id}/messages", a.adminAuth(a.handleAdminConversationMessages))
	mux.HandleFunc("GET /api/admin/announcements", a.adminAuth(a.handleAdminListAnnouncements))
	mux.HandleFunc("POST /api/admin/announcements", a.adminAuth(a.handleAdminCreateAnnouncement))
	mux.HandleFunc("PUT /api/admin/announcements/{id}", a.adminAuth(a.handleAdminUpdateAnnouncement))
	mux.HandleFunc("DELETE /api/admin/announcements/{id}", a.adminAuth(a.handleAdminDeleteAnnouncement))
	mux.HandleFunc("GET /api/admin/settings", a.adminAuth(a.handleAdminGetSettings))
	mux.HandleFunc("PUT /api/admin/settings", a.adminAuth(a.handleAdminSetSettings))
	mux.HandleFunc("GET /api/admin/ai/status", a.adminAuth(a.handleAdminAIStatus))

	return logMiddleware(corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The pattern mux answers matched routes (and 405s wrong methods on
		// known paths); anything outside /api/* falls to the JSON 404 handler.
		if strings.HasPrefix(r.URL.Path, "/api/") {
			mux.ServeHTTP(w, r)
			return
		}
		notFoundHandler(w, r)
	})))
}

// corsMiddleware mirrors the Node cors() package defaults: wildcard origin,
// standard methods/headers, 204 preflight — so non-same-origin deployments
// keep working exactly as they did against the Node backend.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "GET,HEAD,PUT,PATCH,POST,DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// notFoundHandler returns Node-style JSON 404s for unmatched paths instead of
// the mux's plain-text default.
func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	_ = json.NewEncoder(w).Encode(map[string]any{"message": "接口不存在"})
}
