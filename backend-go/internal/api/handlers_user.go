package api

import (
	"net/http"

	"ai-chat-backend/internal/model"
)

// ---------- auth ----------

func (a *API) handleRegister(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Phone    string `json:"phone"`
		Nickname string `json:"nickname"`
		Password string `json:"password"`
	}
	if err := decodeJSON(w, r, &body); err != nil {
		failMsg(w, 400, "请求参数错误")
		return
	}
	if serr := a.Svc.Register(body.Phone, body.Nickname, body.Password); serr != nil {
		fail(w, serr)
		return
	}
	ok(w, true)
}

func (a *API) handleLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Phone    string `json:"phone"`
		Password string `json:"password"`
	}
	if err := decodeJSON(w, r, &body); err != nil {
		failMsg(w, 400, "请求参数错误")
		return
	}
	token, serr := a.Svc.Login(body.Phone, body.Password)
	if serr != nil {
		fail(w, serr)
		return
	}
	ok(w, map[string]any{"token": token})
}

func (a *API) handleLogout(user *model.User, w http.ResponseWriter, r *http.Request) {
	if serr := a.Svc.Logout(user.ID); serr != nil {
		fail(w, serr)
		return
	}
	ok(w, true)
}

// ---------- user ----------

func handleSessionCheck(_ *model.User, w http.ResponseWriter, _ *http.Request) {
	ok(w, map[string]any{"ok": true})
}

func (a *API) handleGetProfile(user *model.User, w http.ResponseWriter, _ *http.Request) {
	ok(w, user.SafeUser())
}

func (a *API) handleUpdateProfile(user *model.User, w http.ResponseWriter, r *http.Request) {
	var body struct {
		Nickname  string `json:"nickname"`
		AvatarURL string `json:"avatarUrl"`
	}
	if err := decodeJSON(w, r, &body); err != nil {
		failMsg(w, 400, "请求参数错误")
		return
	}
	updated, serr := a.Svc.UpdateProfile(user.ID, body.Nickname, body.AvatarURL)
	if serr != nil {
		fail(w, serr)
		return
	}
	ok(w, updated.SafeUser())
}

func (a *API) handleChangePassword(user *model.User, w http.ResponseWriter, r *http.Request) {
	var body struct {
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}
	if err := decodeJSON(w, r, &body); err != nil {
		failMsg(w, 400, "请求参数错误")
		return
	}
	if serr := a.Svc.ChangePassword(user.ID, body.OldPassword, body.NewPassword); serr != nil {
		fail(w, serr)
		return
	}
	ok(w, true)
}

func (a *API) handlePreferences(user *model.User, w http.ResponseWriter, r *http.Request) {
	var body struct {
		AnswerLength string `json:"answerLength"`
		AnswerStyle  string `json:"answerStyle"`
	}
	if err := decodeJSON(w, r, &body); err != nil {
		failMsg(w, 400, "请求参数错误")
		return
	}
	if serr := a.Svc.UpdatePreferences(user.ID, body.AnswerLength, body.AnswerStyle); serr != nil {
		fail(w, serr)
		return
	}
	ok(w, true)
}

func (a *API) handleRedeem(user *model.User, w http.ResponseWriter, r *http.Request) {
	var body struct {
		Code string `json:"code"`
	}
	if err := decodeJSON(w, r, &body); err != nil {
		failMsg(w, 400, "请求参数错误")
		return
	}
	updated, serr := a.Svc.Redeem(user.ID, body.Code)
	if serr != nil {
		fail(w, serr)
		return
	}
	ok(w, map[string]any{"profile": updated.SafeUser()})
}

// ---------- announcements (user side) ----------

func (a *API) handleCurrentAnnouncement(user *model.User, w http.ResponseWriter, _ *http.Request) {
	ann, err := a.Svc.Store.CurrentUnreadAnnouncement(user.ID)
	if err != nil {
		ok(w, nil)
		return
	}
	ok(w, map[string]any{
		"id":        ann.ID,
		"title":     ann.Title,
		"content":   ann.Content,
		"createdAt": ann.CreatedAt,
	})
}

func (a *API) handleAckAnnouncement(user *model.User, w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := a.Svc.Store.AckAnnouncement(id, user.ID); err != nil {
		failMsg(w, 500, "服务器繁忙")
		return
	}
	ok(w, true)
}
