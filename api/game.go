package api

import (
	"net/http"

	"github.com/faria/traveling-hub/internal/agent"
	appplatform "github.com/faria/traveling-hub/internal/platform/app"
)

func registerGameRoutes(mux *http.ServeMux, application *appplatform.App) {
	mux.Handle("GET /v1/me/game", authenticated(application, func(w http.ResponseWriter, r *http.Request, principal agent.Principal) {
		frog, err := application.Frogs.GetForAgent(r.Context(), principal.AgentID)
		if err != nil {
			writeError(w, http.StatusNotFound, "agent traveler not found")
			return
		}
		snapshot, err := application.Journeys.ReconcileAndSnapshot(r.Context(), frog.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "unable to load traveler state")
			return
		}
		writeJSON(w, http.StatusOK, snapshot)
	}))

	mux.Handle("GET /v1/game", webSession(application, false, func(w http.ResponseWriter, r *http.Request, _ string, session webPrincipal) {
		frog, err := application.Frogs.GetForUser(r.Context(), session.UserID)
		if err != nil {
			writeError(w, http.StatusNotFound, "user traveler not found")
			return
		}
		snapshot, err := application.Journeys.ReconcileAndSnapshot(r.Context(), frog.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "unable to load traveler state")
			return
		}
		writeJSON(w, http.StatusOK, snapshot)
	}))
}
