package middleware

import (
	"net/http"

	"github.com/gamedb/gamedb/pkg/session"
)

func MiddlewareAuthCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if session.IsLoggedIn(r) {
			next.ServeHTTP(w, r)
			return
		}

		// session.SetFlash(r, session.SessionBad, "Please login")
		// session.Save(w, r)

		http.Redirect(w, r, "/login", http.StatusFound)
	})
}

func MiddlewareAdminCheck(errorHandler http.HandlerFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if session.IsAdmin(r) {
				next.ServeHTTP(w, r)
				return
			}

			errorHandler(w, r)
		})
	}
}
