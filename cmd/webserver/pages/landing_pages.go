package pages

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi"
)

const (
	LandingAPI      = "/steam-api"
	LandingDeals    = "/top-steam-deals"
	LandingTopGames = "/top-steam-games"
	LandingXP       = "/steam-xp-table"
)

func LandingPagesRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/{id}", landingPagesHandler)
	return r
}

func landingPagesHandler(w http.ResponseWriter, r *http.Request) {

	t := GlobalTemplate{}
	t.fill(w, r, "Info", "")
	t.hideAds = true

	switch strings.Replace(r.URL.Path, "/lp", "", 1) {
	case LandingAPI:
		returnTemplate(w, r, "landing_api", t)
	case LandingDeals:
		returnTemplate(w, r, "landing_deals", t)
	case LandingTopGames:
		returnTemplate(w, r, "landing_games", t)
	case LandingXP:
		returnTemplate(w, r, "landing_xp", t)
	default:
		returnErrorTemplate(w, r, errorTemplate{Code: 404})
		return
	}
}
