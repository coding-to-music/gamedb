package middleware

import (
	"net/http"
	"time"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/ratelimit"
)

func RateLimiterBlock(per time.Duration, burst int, block bool) func(http.Handler) http.Handler {

	limiters := helpers.NewLimiters(per, burst)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if block {
				err := limiters.GetLimiter(r.RemoteAddr).Wait(r.Context())
				if err != nil {
					log.ErrS(err)
					http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
					return
				}
			} else {
				if !limiters.GetLimiter(r.RemoteAddr).Allow() {
					http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
