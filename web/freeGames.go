package web

import (
	"net/http"
	"net/url"

	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
)

func FreeGamesHandler(w http.ResponseWriter, r *http.Request) {

	search := url.Values{}
	search.Set("is_free", "1")
	search.Set("name", "-")
	search.Set("type", "game")

	// Types not in this list will show first
	sort := "FIELD(`type`,'game','dlc','demo','mod','video','movie','series','episode','application','tool','advertising'), name ASC"
	freeApps, err := mysql.SearchApps(search, 1000, sort, []string{"id", "name", "icon", "type", "platforms"})
	if err != nil {
		logger.Error(err)
	}

	template := freeGamesTemplate{}
	template.Fill(r, "Free Games")
	template.Apps = freeApps

	returnTemplate(w, r, "free_games", template)
	return
}

type freeGamesTemplate struct {
	GlobalTemplate
	Apps []mysql.App
}
