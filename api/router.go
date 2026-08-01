package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/faria/traveling-hub/internal/agent"
	"github.com/faria/traveling-hub/internal/event"
	appplatform "github.com/faria/traveling-hub/internal/platform/app"
)

func NewRouter(application *appplatform.App) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := application.Healthy(r.Context()); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": application.Config.BuildVersion})
	})
	mux.Handle("GET /v1/me", authenticated(application, func(w http.ResponseWriter, r *http.Request, principal agent.Principal) {
		frog, err := application.Frogs.GetForAgent(r.Context(), principal.AgentID)
		if err != nil {
			writeError(w, http.StatusNotFound, "agent traveler not found")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"agent_id": principal.AgentID.String(), "frog_id": frog.ID.String()})
	}))
	mux.Handle("GET /v1/me/events", authenticated(application, func(w http.ResponseWriter, r *http.Request, principal agent.Principal) {
		limit, err := event.ParseLimit(r.URL.Query().Get("limit"))
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		events, next, err := application.Events.ListAfter(r.Context(), principal.AgentID, r.URL.Query().Get("cursor"), limit)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"events": events, "next_cursor": next})
	}))
	mux.Handle("POST /v1/me/events/ack", authenticated(application, func(w http.ResponseWriter, r *http.Request, principal agent.Principal) {
		var request struct {
			Cursor string `json:"cursor"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
		if err := application.Events.Acknowledge(r.Context(), principal.AgentID, request.Cursor); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	if application.Config.Environment == "development" {
		mux.Handle("POST /v1/dev/fixture-tick", authenticated(application, func(w http.ResponseWriter, r *http.Request, principal agent.Principal) {
			if _, err := application.Frogs.GetForAgent(r.Context(), principal.AgentID); err != nil {
				writeError(w, http.StatusNotFound, "agent traveler not found")
				return
			}
			result, err := application.Simulation.RunFixtureTick(r.Context(), principal.AgentID)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "unable to run fixture tick")
				return
			}
			writeJSON(w, http.StatusCreated, result)
		}))
	}
	registerIdentityRoutes(mux, application)
	registerWebAuthRoutes(mux, application)
	registerGameRoutes(mux, application)
	return mux
}

type authenticatedHandler func(http.ResponseWriter, *http.Request, agent.Principal)

func authenticated(application *appplatform.App, next authenticatedHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := strings.TrimSpace(r.Header.Get("Authorization"))
		if !strings.HasPrefix(header, "Bearer ") || strings.TrimSpace(strings.TrimPrefix(header, "Bearer ")) == "" {
			writeError(w, http.StatusUnauthorized, "missing or invalid authorization header")
			return
		}
		principal, err := application.Agents.Authenticate(r.Context(), strings.TrimSpace(strings.TrimPrefix(header, "Bearer ")))
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unable to authenticate agent")
			return
		}
		next(w, r, principal)
	})
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
