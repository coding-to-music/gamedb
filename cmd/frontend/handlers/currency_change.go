package handlers

import (
	"net/http"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/session"
	"github.com/gamedb/gamedb/pkg/i18n"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/go-chi/chi"
)

func CurrencyHandler(w http.ResponseWriter, r *http.Request) {

	id := steamapi.ProductCC(chi.URLParam(r, "id"))
	if id == "" {
		id = steamapi.ProductCCUS
	}

	if i18n.IsValidProdCC(id) {

		// Set to session
		session.Set(r, session.SessionUserProdCC, string(id))

		// Set to user row
		user, err := getUserFromSession(r)
		if err == nil {
			err = user.SetProdCC(id)
			if err != nil {
				log.ErrS(err)
			}
		}

	} else {
		session.SetFlash(r, session.SessionGood, "Invalid currency")
	}

	// Save session
	session.Save(w, r)

	// Redirect
	lastPage := session.Get(r, session.SessionLastPage)

	if lastPage == "" {
		lastPage = "/"
	}

	http.Redirect(w, r, lastPage, http.StatusFound)
}
