package middleware

import (
	"net/http"
	"strings"

	"github.com/gamedb/gamedb/pkg/log"
)

var DownMessage string

func MiddlewareDownMessage(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if DownMessage == "" || strings.HasPrefix(r.URL.Path, "/admin") {
			next.ServeHTTP(w, r)
		} else {
			_, err := w.Write([]byte(DownMessage))
			if err != nil {
				log.ErrS(err)
			}
		}
	})
}
