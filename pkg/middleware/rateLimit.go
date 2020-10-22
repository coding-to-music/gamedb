package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gamedb/gamedb/pkg/log"
	"golang.org/x/time/rate"
)

type ipLimiters struct {
	ips   map[string]*ipLimiter
	lock  sync.Mutex
	limit rate.Limit
	burst int
}

type ipLimiter struct {
	limiter *rate.Limiter
	updated time.Time
}

func (l *ipLimiters) GetLimiter(ip string) *rate.Limiter {

	limiter, exists := l.ips[ip]

	if !exists {

		limiter = &ipLimiter{
			limiter: rate.NewLimiter(l.limit, l.burst),
		}

		l.lock.Lock()
		l.ips[ip] = limiter
		l.lock.Unlock()
	}

	// Touch IP
	limiter.updated = time.Now()

	return limiter.limiter
}

func (l *ipLimiters) Clean() {
	for {
		cutoff := time.Now().Add(time.Hour * -1)

		l.lock.Lock()
		for k, v := range l.ips {
			if v.updated.Before(cutoff) {
				delete(l.ips, k)
			}
		}
		l.lock.Unlock()

		time.Sleep(time.Minute)
	}
}

func newLimiters(per time.Duration, burst int) *ipLimiters {

	if burst < 1 {
		burst = 1
	}

	l := &ipLimiters{
		ips:   map[string]*ipLimiter{},
		lock:  sync.Mutex{},
		limit: rate.Every(per),
		burst: burst,
	}

	go l.Clean()

	return l
}

func RateLimiterBlock(per time.Duration, burst int, block bool) func(http.Handler) http.Handler {

	limiters := newLimiters(per, burst)

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
