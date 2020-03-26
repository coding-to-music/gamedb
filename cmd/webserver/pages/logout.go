package pages

import (
	"net/http"

	"github.com/Jleagle/session-go/session"
	webserverHelpers "github.com/gamedb/gamedb/cmd/webserver/helpers"
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
	userID, err := webserverHelpers.GetUserIDFromSesion(r)
	if err != nil {
		log.Err(err, r)
	} else {
		err = mongo.CreateUserEvent(r, userID, mongo.EventLogout)
		log.Err(err, r)
	}

	// Get last page
	lastPage, err := session.Get(r, webserverHelpers.SessionLastPage)
	if err != nil {
		log.Err(err, r)
	}

	if lastPage == "" {
		lastPage = "/"
	}

	// Logout
	err = session.DeleteAll(r)
	log.Err(err, r)

	err = session.SetFlash(r, webserverHelpers.SessionGood, "You have been logged out")
	log.Err(err, r)

	err = session.Save(w, r)
	log.Err(err, r)

	//
	http.Redirect(w, r, lastPage, http.StatusFound)
}
