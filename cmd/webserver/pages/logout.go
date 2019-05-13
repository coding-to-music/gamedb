package pages

import (
	"net/http"

	"github.com/gamedb/gamedb/cmd/webserver/session"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
)

func LogoutRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", logoutHandler)
	return r
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	// Make event
	steamID, err := getPlayerIDFromSession(r)
	if err != nil {
		err = helpers.IgnoreErrors(err, eeeNoPlayerIDSet)
		log.Err(err, r)
	} else {
		err = mongo.CreateEvent(r, steamID, mongo.EventLogout)
		log.Err(err, r)
	}

	// Logout
	err = session.Clear(r)
	log.Err(err, r)

	err = session.SetGoodFlash(r, "You have been logged out")
	log.Err(err, r)

	err = session.Save(w, r)
	log.Err(err, r)

	//
	http.Redirect(w, r, "/", http.StatusFound)
}
