package middleware

import (
	"net/http"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/go-chi/cors"
)

// todo, check this is alright
func MiddlewareCors() func(next http.Handler) http.Handler {
	return cors.New(cors.Options{
		AllowedOrigins: []string{config.C.GameDBDomain}, // Use this to allow specific origin hosts
		AllowedMethods: []string{"GET", "POST"},
	}).Handler
}
