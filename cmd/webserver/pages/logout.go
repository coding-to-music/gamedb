package pages

import (
	"net/http"

	"github.com/Jleagle/session-go/session"
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

	// Make event
	userID, err := helpers.GetUserIDFromSesion(r)
	if err != nil || userID == 0 {
		log.Err(err, r)
	} else {
		err = mongo.CreateUserEvent(r, userID, mongo.EventLogout)
		log.Err(err, r)
	}

	// Logout
	err = session.DeleteAll(r)
	log.Err(err, r)

	err = session.SetFlash(r, helpers.SessionGood, "You have been logged out")
	log.Err(err, r)

	err = session.Save(w, r)
	log.Err(err, r)

	// Get last page
	val, err := session.Get(r, helpers.SessionLastPage)
	if err != nil {
		log.Err(err, r)
	}

	if val == "" {
		val = "/"
	}

	//
	http.Redirect(w, r, val, http.StatusFound)
}
