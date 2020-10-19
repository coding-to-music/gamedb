package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gamedb/gamedb/cmd/frontend/helpers/session"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/go-chi/cors"
	"github.com/justinas/nosurf"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
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

		cf := r.Header.Get("cf-connecting-ip")
		nginx := r.Header.Get("x-real-ip")

		if cf != "" {
			r.RemoteAddr = cf
		} else if nginx != "" {
			r.RemoteAddr = nginx
		}

		h.ServeHTTP(w, r)
	})
}

var (
	limit    = rate.NewLimiter(rate.Every(time.Second), 10)
	limitCtx = context.TODO()
)

func NewDelayingLimiter(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		err := limit.Wait(limitCtx)
		if err != nil {

			w.WriteHeader(429)
			_, err := w.Write([]byte("rate limited"))
			if err != nil {
				log.ErrS(err)
			}
			return
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
			if err != nil {
				log.ErrS(err)
			}
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

			// session.SetFlash(r, session.SessionBad, "Please login")
			// session.Save(w, r)

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
