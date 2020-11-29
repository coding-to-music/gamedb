package middleware

import (
	"net/http"
	"regexp"

	"github.com/gamedb/gamedb/pkg/helpers"
)

var ipSplitRegex = regexp.MustCompile("[:,]")

func MiddlewareRealIP(next http.Handler) http.Handler {
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

		r.RemoteAddr = ipSplitRegex.Split(r.RemoteAddr, 2)[0]

		next.ServeHTTP(w, r)
	})
}
