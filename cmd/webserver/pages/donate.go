package pages

import (
	"net/http"

	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/go-chi/chi"
)

func DonateRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", donateHandler)
	return r
}

func donateHandler(w http.ResponseWriter, r *http.Request) {

	t := donateTemplate{}
	t.fill(w, r, "Donate", "Databases take up a tonne of resources. Help pay for the server costs or just buy me a beer.")
	t.Pages = map[int]int{
		sql.UserLevel0: sql.UserLevelLimit0,
		sql.UserLevel1: sql.UserLevelLimit1,
		sql.UserLevel2: sql.UserLevelLimit2,
		sql.UserLevel3: sql.UserLevelLimit3,
	}

	returnTemplate(w, r, "donate", t)
}

type donateTemplate struct {
	GlobalTemplate
	Pages map[int]int
}
