package api

import (
	"net/http"
	"strings"

	"ai-chat-backend/internal/model"
)

func lower(s string) string { return strings.ToLower(s) }

func contains(haystack, needle string) bool {
	return needle != "" && strings.Contains(haystack, needle)
}

// ---------- admin auth ----------

func (a *API) handleAdminLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	token, serr := a.Svc.AdminLogin(body.Username, body.Password)
	if serr != nil {
		fail(w, serr)
		return
	}
	ok(w, map[string]any{"token": token})
}

func (a *API) handleAdminChangePassword(admin *model.User, w http.ResponseWriter, r *http.Request) {
	var body struct {
		NewPassword string `json:"newPassword"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	if serr := a.Svc.AdminChangePassword(admin.ID, body.NewPassword); serr != nil {
		fail(w, serr)
		return
	}
	ok(w, true)
}

// ---------- dashboard / users ----------

func (a *API) handleAdminDashboard(_ *model.User, w http.ResponseWriter, _ *http.Request) {
	data, serr := a.Svc.Dashboard()
	if serr != nil {
		fail(w, serr)
		return
	}
	ok(w, data)
}

func (a *API) handleAdminUsers(_ *model.User, w http.ResponseWriter, r *http.Request) {
	result, serr := a.Svc.AdminUsers(queryInt(r, "page", 1), queryInt(r, "pageSize", 10), queryString(r, "phone"), queryString(r, "status"))
	if serr != nil {
		fail(w, serr)
		return
	}
	ok(w, result)
}

func (a *API) handleAdminUpdateUser(_ *model.User, w http.ResponseWriter, r *http.Request) {
	var body struct {
		Nickname string `json:"nickname"`
		Phone    string `json:"phone"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	if serr := a.Svc.AdminUpdateUser(r.PathValue("id"), body.Nickname, body.Phone); serr != nil {
		fail(w, serr)
		return
	}
	ok(w, true)
}

func (a *API) handleAdminUserStatus(_ *model.User, w http.ResponseWriter, r *http.Request) {
	var body struct {
		Status string `json:"status"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	if serr := a.Svc.AdminSetUserStatus(r.PathValue("id"), body.Status, "web-admin"); serr != nil {
		fail(w, serr)
		return
	}
	ok(w, true)
}

func (a *API) handleAdminResetPassword(_ *model.User, w http.ResponseWriter, r *http.Request) {
	var body struct {
		NewPassword string `json:"newPassword"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	if serr := a.Svc.AdminResetPassword(r.PathValue("id"), body.NewPassword, "web-admin"); serr != nil {
		fail(w, serr)
		return
	}
	ok(w, true)
}

func (a *API) handleAdminMemberExpire(_ *model.User, w http.ResponseWriter, r *http.Request) {
	var body struct {
		MemberExpireAt *string `json:"memberExpireAt"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	value := ""
	if body.MemberExpireAt != nil {
		value = *body.MemberExpireAt
	}
	updated, serr := a.Svc.AdminSetMemberExpire(r.PathValue("id"), value, "web-admin")
	if serr != nil {
		fail(w, serr)
		return
	}
	ok(w, updated.SafeUser())
}

func (a *API) handleAdminMembers(_ *model.User, w http.ResponseWriter, _ *http.Request) {
	members, serr := a.Svc.AdminMembers()
	if serr != nil {
		fail(w, serr)
		return
	}
	ok(w, members)
}

// ---------- roles / menus ----------

func (a *API) handleAdminRoles(_ *model.User, w http.ResponseWriter, _ *http.Request) {
	roles, err := a.Svc.Store.GetRoles()
	if err != nil {
		failMsg(w, 500, "服务器繁忙")
		return
	}
	ok(w, roles)
}

func (a *API) handleAdminMenus(_ *model.User, w http.ResponseWriter, _ *http.Request) {
	menus, err := a.Svc.Store.GetMenus()
	if err != nil {
		failMsg(w, 500, "服务器繁忙")
		return
	}
	ok(w, menus)
}

// ---------- redeem codes / records ----------

func (a *API) handleAdminRedeemCodes(_ *model.User, w http.ResponseWriter, r *http.Request) {
	result, serr := a.Svc.AdminRedeemCodes(queryInt(r, "page", 1), queryInt(r, "pageSize", 10), queryString(r, "status"), queryString(r, "code"))
	if serr != nil {
		fail(w, serr)
		return
	}
	ok(w, result)
}

func (a *API) handleAdminRedeemCodesExport(_ *model.User, w http.ResponseWriter, r *http.Request) {
	list, serr := a.Svc.AdminRedeemCodesExport(queryString(r, "status"), queryString(r, "code"))
	if serr != nil {
		fail(w, serr)
		return
	}
	ok(w, list)
}

func (a *API) handleAdminRedeemCodesBatch(_ *model.User, w http.ResponseWriter, r *http.Request) {
	var body struct {
		Quantity       *int `json:"quantity"`
		DurationMonths *int `json:"durationMonths"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	quantity := 5
	duration := 1
	if body.Quantity != nil {
		quantity = *body.Quantity
	}
	if body.DurationMonths != nil {
		duration = *body.DurationMonths
	}
	created, serr := a.Svc.GenerateRedeemCodes(quantity, duration, "web-admin")
	if serr != nil {
		fail(w, serr)
		return
	}
	ok(w, created)
}

func (a *API) handleAdminRedeemCodeVoid(_ *model.User, w http.ResponseWriter, r *http.Request) {
	if serr := a.Svc.AdminVoidRedeemCode(r.PathValue("id"), "web-admin"); serr != nil {
		fail(w, serr)
		return
	}
	ok(w, true)
}

func (a *API) handleAdminRedeemRecords(_ *model.User, w http.ResponseWriter, r *http.Request) {
	result, serr := a.Svc.AdminRedeemRecords(queryInt(r, "page", 1), queryInt(r, "pageSize", 10), queryString(r, "phone"), queryString(r, "start"), queryString(r, "end"))
	if serr != nil {
		fail(w, serr)
		return
	}
	ok(w, result)
}

// ---------- conversations ----------

func (a *API) handleAdminConversations(_ *model.User, w http.ResponseWriter, r *http.Request) {
	page := queryInt(r, "page", 1)
	pageSize := queryInt(r, "pageSize", 10)
	phone := queryString(r, "phone")
	keyword := queryString(r, "keyword")
	search := queryString(r, "search")

	result, serr := a.Svc.AdminConversations(page, pageSize, phone, keyword, search)
	if serr != nil {
		fail(w, serr)
		return
	}
	ok(w, result)
}

func (a *API) handleAdminConversationMessages(_ *model.User, w http.ResponseWriter, r *http.Request) {
	msgs, serr := a.Svc.AdminConversationMessages(r.PathValue("id"))
	if serr != nil {
		fail(w, serr)
		return
	}
	ok(w, msgs)
}
