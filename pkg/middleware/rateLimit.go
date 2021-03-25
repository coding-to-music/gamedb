package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Jleagle/rate-limit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	xrate "golang.org/x/time/rate"
)

func SetRateLimitHeaders(w http.ResponseWriter, limiters *rate.Limiters, reservation *xrate.Reservation) {

	w.Header().Set("X-RateLimit-Every", limiters.GetMinInterval().String())
	w.Header().Set("X-RateLimit-Burst", fmt.Sprint(limiters.GetBurst()))
	w.Header().Set("X-RateLimit-Wait", reservation.Delay().String())
	w.Header().Set("X-RateLimit-Bucket", "global")
}

func RateLimiterBlock(limiters *rate.Limiters, handler http.HandlerFunc) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			reservation := limiters.GetLimiter(r.RemoteAddr).Reserve()

			SetRateLimitHeaders(w, limiters, reservation)

			if !reservation.OK() {
				handler(w, r)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RateLimiterWait(limiters *rate.Limiters) func(http.Handler) http.Handler {

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
