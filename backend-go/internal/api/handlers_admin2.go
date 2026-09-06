package api

import (
	"net/http"

	"ai-chat-backend/internal/model"
)

// ---------- announcements (admin) ----------

func (a *API) handleAdminListAnnouncements(_ *model.User, w http.ResponseWriter, _ *http.Request) {
	list, serr := a.Svc.AdminListAnnouncements()
	if serr != nil {
		fail(w, serr)
		return
	}
	ok(w, list)
}

func (a *API) handleAdminCreateAnnouncement(_ *model.User, w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		Active  *bool  `json:"active"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	active := true
	if body.Active != nil {
		active = *body.Active
	}
	ann, serr := a.Svc.AdminCreateAnnouncement(body.Title, body.Content, active, "web-admin")
	if serr != nil {
		fail(w, serr)
		return
	}
	ok(w, ann)
}

func (a *API) handleAdminUpdateAnnouncement(_ *model.User, w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		Active  *bool  `json:"active"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	if serr := a.Svc.AdminUpdateAnnouncement(r.PathValue("id"), body.Title, body.Content, body.Active, "web-admin"); serr != nil {
		fail(w, serr)
		return
	}
	ok(w, true)
}

func (a *API) handleAdminDeleteAnnouncement(_ *model.User, w http.ResponseWriter, r *http.Request) {
	if serr := a.Svc.AdminDeleteAnnouncement(r.PathValue("id")); serr != nil {
		fail(w, serr)
		return
	}
	ok(w, true)
}

// ---------- settings ----------

func (a *API) handleAdminGetSettings(_ *model.User, w http.ResponseWriter, _ *http.Request) {
	ok(w, map[string]any{
		"registrationOpen": a.Svc.RegistrationOpen(),
		"redeemOpen":       a.Svc.RedeemOpen(),
	})
}

func (a *API) handleAdminSetSettings(_ *model.User, w http.ResponseWriter, r *http.Request) {
	var body struct {
		RegistrationOpen *bool `json:"registrationOpen"`
		RedeemOpen       *bool `json:"redeemOpen"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	if body.RegistrationOpen != nil {
		if err := a.Svc.SetRegistrationOpen(*body.RegistrationOpen, "web-admin"); err != nil {
			failMsg(w, 500, "服务器繁忙")
			return
		}
	}
	if body.RedeemOpen != nil {
		if err := a.Svc.SetRedeemOpen(*body.RedeemOpen, "web-admin"); err != nil {
			failMsg(w, 500, "服务器繁忙")
			return
		}
	}
	ok(w, true)
}

// ---------- AI status ----------

func (a *API) handleAdminAIStatus(_ *model.User, w http.ResponseWriter, _ *http.Request) {
	ok(w, map[string]any{
		"mode":   a.Router.Mode(),
		"models": a.Router.HealthSnapshot(),
		"stats":  a.Router.StatsSnapshot(),
	})
}
