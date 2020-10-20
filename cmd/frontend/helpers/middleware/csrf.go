package middleware

import (
	"net/http"

	"github.com/justinas/nosurf"
)

func MiddlewareCSRF(h http.Handler) http.Handler {
	return nosurf.New(h)
}
