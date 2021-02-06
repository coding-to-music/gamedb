package handlers

import (
	"net/http"

	"github.com/gamedb/gamedb/cmd/frontend/helpers/session"
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
	userID := session.GetUserIDFromSesion(r)
	if userID > 0 {
		err := mongo.NewEvent(r, userID, mongo.EventLogout)
		if err != nil {
			log.ErrS(err)
		}
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
