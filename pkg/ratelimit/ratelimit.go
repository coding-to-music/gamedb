package ratelimit

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type limiters struct {
	limiters map[string]*limiterInner
	lock     sync.Mutex
	limit    rate.Limit
	burst    int
}

type limiterInner struct {
	limiter *rate.Limiter
	updated time.Time
}

func (l *limiters) GetLimiter(key string) *rate.Limiter {

	limiter, exists := l.limiters[key]

	if !exists {

		limiter = &limiterInner{
			limiter: rate.NewLimiter(l.limit, l.burst),
		}

		l.lock.Lock()
		l.limiters[key] = limiter
		l.lock.Unlock()
	}

	// Touch limiter
	limiter.updated = time.Now()

	return limiter.limiter
}

func (l *limiters) clean() {

	for {
		cutoff := time.Now().Add(time.Hour * -1)

		l.lock.Lock()
		for k, v := range l.limiters {
			if v.updated.Before(cutoff) {
				delete(l.limiters, k)
			}
		}
		l.lock.Unlock()

		time.Sleep(time.Minute)
	}
}

func New(per time.Duration, burst int) *limiters {

	if burst < 1 {
		burst = 1
	}

	l := &limiters{
		limiters: map[string]*limiterInner{},
		lock:     sync.Mutex{},
		limit:    rate.Every(per),
		burst:    burst,
	}

	go l.clean()

	return l
}
