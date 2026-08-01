package api

import (
	"net/http"
	"strings"

	"github.com/faria/traveling-hub/internal/identity"
	appplatform "github.com/faria/traveling-hub/internal/platform/app"
)

type webPrincipal struct{ identity.UserSession }
type webHandler func(http.ResponseWriter, *http.Request, string, webPrincipal)

func webSession(application *appplatform.App, allowRestricted bool, next webHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(webSessionCookieName)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "web session required")
			return
		}
		session, err := application.Identity.LoadSession(r.Context(), cookie.Value)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "web session required")
			return
		}
		if session.MustChange && !allowRestricted {
			writeError(w, http.StatusForbidden, "password change required")
			return
		}
		if isUnsafeOrigin(r, application.Config.WebOrigin) {
			writeError(w, http.StatusForbidden, "invalid request origin")
			return
		}
		next(w, r, cookie.Value, webPrincipal{UserSession: session})
	})
}

func isUnsafeOrigin(r *http.Request, allowedOrigin string) bool {
	if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
		return false
	}
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	return origin == "" || origin != allowedOrigin
}
