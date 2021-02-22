package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/ratelimit"
)

func RateLimiterBlock(per time.Duration, burst int, handler http.HandlerFunc) func(http.Handler) http.Handler {

	limiters := ratelimit.New(per, burst)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			reservation := limiters.GetLimiter(r.RemoteAddr).Reserve()

			w.Header().Set("X-RateLimit-Every", per.String())
			w.Header().Set("X-RateLimit-Burst", fmt.Sprint(burst))
			w.Header().Set("X-RateLimit-Wait", reservation.Delay().String())
			w.Header().Set("X-RateLimit-Bucket", "global")

			if !reservation.OK() {
				handler(w, r)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RateLimiterWait(per time.Duration, burst int) func(http.Handler) http.Handler {

	limiters := ratelimit.New(per, burst)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			err := limiters.GetLimiter(r.RemoteAddr).Wait(r.Context())
			if err != nil {
				err = helpers.IgnoreErrors(err, context.Canceled)
				if err != nil {
					log.ErrS(err)
				}
				http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
