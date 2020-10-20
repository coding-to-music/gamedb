package middleware

import (
	"net/http"
)

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
