package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gamedb/gamedb/cmd/frontend/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/go-chi/cors"
	"github.com/justinas/nosurf"
	"go.uber.org/zap"
)

func MiddlewareCSRF(h http.Handler) http.Handler {
	return nosurf.New(h)
}

// todo, check this is alright
func MiddlewareCors() func(next http.Handler) http.Handler {
	return cors.New(cors.Options{
		AllowedOrigins: []string{config.C.GameDBDomain}, // Use this to allow specific origin hosts
		AllowedMethods: []string{"GET", "POST"},
	}).Handler
}

func MiddlewareRealIP(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		rip := r.Header.Get("X-Real-IP")
		if rip != "" {
			r.RemoteAddr = rip
		}
		h.ServeHTTP(w, r)
	})
}

var DownMessage string

func MiddlewareDownMessage(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if DownMessage == "" || strings.HasPrefix(r.URL.Path, "/admin") {
			h.ServeHTTP(w, r)
		} else {
			_, err := w.Write([]byte(DownMessage))
			zap.S().Error(err)
		}
	})
}

func MiddlewareTime(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		r.Header.Set("start-time", strconv.FormatInt(time.Now().UnixNano(), 10))

		next.ServeHTTP(w, r)
	})
}

func MiddlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if config.IsLocal() {
			zap.S().Named(log.LogNameRequests).Info(r.Method + " " + r.URL.String())
		}
		next.ServeHTTP(w, r)
	})
}

func MiddlewareAuthCheck() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if session.IsLoggedIn(r) {
				next.ServeHTTP(w, r)
				return
			}

			session.SetFlash(r, session.SessionBad, "Please login")
			session.Save(w, r)

			http.Redirect(w, r, "/login", http.StatusFound)
		})
	}
}

func MiddlewareAdminCheck(handler http.HandlerFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if session.IsAdmin(r) {
				next.ServeHTTP(w, r)
				return
			}

			handler(w, r)
		})
	}
}
