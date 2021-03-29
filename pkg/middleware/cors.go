package middleware

import (
	"net/http"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/go-chi/cors"
)

func MiddlewareCors() func(next http.Handler) http.Handler {

	options := cors.Options{
		AllowedMethods: []string{"GET", "POST"},
	}

	if !config.IsLocal() {
		options.AllowedOrigins = []string{config.C.GlobalSteamDomain, "https://editor.swagger.io"}
	}

	return cors.New(options).Handler
}
