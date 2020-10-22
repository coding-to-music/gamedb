package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"go.uber.org/zap"
)

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
