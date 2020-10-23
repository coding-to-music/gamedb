package email

import (
	"net/http"

	"github.com/gamedb/gamedb/cmd/frontend/helpers/geo"
)

func NewSignup(email string, r *http.Request) error {

	return GetProvider().Send(
		email,
		"",
		"",
		"Welcome to Game DB",
		SignupTemplate{
			IP: geo.GetFirstIP(r.RemoteAddr),
		},
	)
}
