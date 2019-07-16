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
		err = session.Set(r, helpers.SessionUserProdCC, id)
	} else {
		err = session.SetFlash(r, helpers.SessionGood, "Invalid currency")
	}

	if err != nil {
		log.Err(err, r)
	}

	err = session.Save(w, r)
	if err != nil {
		log.Err(err, r)
	}

	val, err := session.Get(r, helpers.SessionLastPage)
	if err != nil {
		log.Err(err, r)
	}

	if val == "" {
		val = "/"
	}

	http.Redirect(w, r, val, http.StatusFound)
}
