package api

import (
	"encoding/json"
	"net/http"

	appplatform "github.com/faria/traveling-hub/internal/platform/app"
)

func registerIdentityRoutes(mux *http.ServeMux, application *appplatform.App) {
	mux.HandleFunc("POST /v1/agent-registrations", func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Email string `json:"email"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
		result, err := application.Identity.Register(r.Context(), request.Email)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid email")
			return
		}
		if !result.Created {
			writeJSON(w, http.StatusAccepted, map[string]string{"status": "registration_already_submitted"})
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{
			"agent_id": result.AgentID.String(), "frog_id": result.FrogID.String(),
			"username": result.Username, "initial_password": result.InitialPassword,
			"agent_api_key": result.AgentAPIKey, "must_change_password": result.MustChange,
		})
	})
}
