package pages

import (
	"net/http"

	"github.com/gamedb/gamedb/cmd/frontend/pages/helpers/session"
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
	userID, err := session.GetUserIDFromSesion(r)
	if err == nil {
		err = mongo.CreateUserEvent(r, userID, mongo.EventLogout)
		log.ErrS(err)
	}

	// Get last page
	lastPage := session.Get(r, session.SessionLastPage)
	if lastPage == "" {
		lastPage = "/"
	}

	// Logout
	session.DeleteAll(r)
	session.SetFlash(r, session.SessionGood, "You have been logged out")
	session.Save(w, r)

	//
	http.Redirect(w, r, lastPage, http.StatusFound)
}
