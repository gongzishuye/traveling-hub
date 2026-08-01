package api

import (
	"encoding/json"
	"net/http"

	appplatform "github.com/faria/traveling-hub/internal/platform/app"
)

const webSessionCookieName = "__Host-travelinghub_session"

func registerWebAuthRoutes(mux *http.ServeMux, application *appplatform.App) {
	mux.HandleFunc("GET /v1/web/verify-email", func(w http.ResponseWriter, r *http.Request) {
		if err := application.Identity.VerifyEmail(r.Context(), r.URL.Query().Get("token")); err != nil {
			writeError(w, http.StatusBadRequest, "invalid or expired verification link")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("POST /v1/web/login", func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
		sessionID, session, err := application.Identity.Login(r.Context(), request.Email, request.Password)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		setSessionCookie(w, sessionID)
		writeJSON(w, http.StatusOK, map[string]bool{"must_change_password": session.MustChange})
	})

	mux.Handle("POST /v1/web/change-password", webSession(application, true, func(w http.ResponseWriter, r *http.Request, sessionID string, session webPrincipal) {
		var request struct {
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
		nextSession, err := application.Identity.ChangePassword(r.Context(), sessionID, session.UserSession, request.Password)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid password")
			return
		}
		setSessionCookie(w, nextSession)
		w.WriteHeader(http.StatusNoContent)
	}))
}

func setSessionCookie(w http.ResponseWriter, sessionID string) {
	http.SetCookie(w, &http.Cookie{
		Name: webSessionCookieName, Value: sessionID, Path: "/", HttpOnly: true,
		Secure: true, SameSite: http.SameSiteLaxMode,
	})
}
