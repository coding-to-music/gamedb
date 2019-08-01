package pages

import (
	"net/http"

	"github.com/Jleagle/session-go/session"
	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/go-chi/chi"
)

func CurrencyHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		id = string(steam.ProductCCUS)
	}

	var err error

	if helpers.IsValidProdCC(steam.ProductCC(id)) {

		// Set to session
		err = session.Set(r, helpers.SessionUserProdCC, id)

		// Set to user row
		user, err := getUserFromSession(r)
		if err == nil {
			user.ProductCC = steam.ProductCC(id)
			err2 := user.Save()
			log.Err(err2)
		}

	} else {
		err = session.SetFlash(r, helpers.SessionGood, "Invalid currency")
	}

	if err != nil {
		log.Err(err, r)
	}

	// Save session
	err = session.Save(w, r)
	if err != nil {
		log.Err(err, r)
	}

	// Redirect
	lastPage, err := session.Get(r, helpers.SessionLastPage)
	if err != nil {
		log.Err(err, r)
	}

	if lastPage == "" {
		lastPage = "/"
	}

	http.Redirect(w, r, lastPage, http.StatusFound)
}
