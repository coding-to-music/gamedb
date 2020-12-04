package middleware

import (
	"net/http"

	"github.com/gamedb/gamedb/pkg/helpers"
)

func RealIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		headers := []string{
			r.Header.Get("CF-Connecting-IP"), // Cloudflare
			r.Header.Get("X-Real-IP"),        // Nginx
			r.Header.Get("X-Client-IP"),      //
			r.Header.Get("X-Forwarded-For"),  //
			r.Header.Get("X-Forwarded"),      // Variation
		}

		for _, v := range headers {
			if v != "" {
				ip := helpers.RegexIP.FindString(v)
				if ip != "" {
					r.RemoteAddr = ip
					break
				}
			}
		}

		r.RemoteAddr = helpers.RegexIP.FindString(r.RemoteAddr)

		next.ServeHTTP(w, r)
	})
}
