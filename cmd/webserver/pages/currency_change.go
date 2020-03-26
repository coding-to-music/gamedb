package pages

import (
	"net/http"

	"github.com/Jleagle/session-go/session"
	"github.com/Jleagle/steam-go/steamapi"
	webserverHelpers "github.com/gamedb/gamedb/cmd/webserver/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/i18n"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/go-chi/chi"
)

func CurrencyHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		id = string(steamapi.ProductCCUS)
	}

	var err error

	if i18n.IsValidProdCC(steamapi.ProductCC(id)) {

		// Set to session
		err = session.Set(r, webserverHelpers.SessionUserProdCC, id)

		// Set to user row
		user, err := getUserFromSession(r)
		if err == nil {
			user.ProductCC = steamapi.ProductCC(id)
			err2 := user.Save()
			log.Err(err2)
		}

	} else {
		err = session.SetFlash(r, webserverHelpers.SessionGood, "Invalid currency")
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
	lastPage, err := session.Get(r, webserverHelpers.SessionLastPage)
	if err != nil {
		log.Err(err, r)
	}

	if lastPage == "" {
		lastPage = "/"
	}

	http.Redirect(w, r, lastPage, http.StatusFound)
}
