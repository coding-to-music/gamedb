package pages

import (
	"net/http"

	"github.com/Jleagle/session-go/session"
	sessionHelpers "github.com/gamedb/gamedb/cmd/webserver/pages/helpers/session"
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
	userID, err := sessionHelpers.GetUserIDFromSesion(r)
	if err == nil {
		err = mongo.CreateUserEvent(r, userID, mongo.EventLogout)
		log.Err(err, r)
	}

	// Get last page
	lastPage := sessionHelpers.Get(r, sessionHelpers.SessionLastPage)
	if lastPage == "" {
		lastPage = "/"
	}

	// Logout
	err = session.DeleteAll(r)
	log.Err(err, r)

	err = session.SetFlash(r, sessionHelpers.SessionGood, "You have been logged out")
	log.Err(err, r)

	sessionHelpers.Save(w, r)

	//
	http.Redirect(w, r, lastPage, http.StatusFound)
}
