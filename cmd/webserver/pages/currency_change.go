package pages

import (
	"net/http"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/i18n"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

func CurrencyHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		id = string(steamapi.ProductCCUS)
	}

	var err error

	if i18n.IsValidProdCC(steamapi.ProductCC(id)) {

		// Set to session
		session.Set(r, session.SessionUserProdCC, id)

		// Set to user row
		user, err := getUserFromSession(r)
		if err == nil {
			user.ProductCC = steamapi.ProductCC(id)
			err2 := user.Save()
			zap.S().Error(err2, r)
		}

	} else {
		session.SetFlash(r, session.SessionGood, "Invalid currency")
		zap.S().Error(err)
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
