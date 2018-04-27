package web

import (
	"net/http"

	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
)

func FreeGamesHandler(w http.ResponseWriter, r *http.Request) {

	db, err := mysql.GetDB()
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 500, "Can't connect to database")
		return
	}

	db = db.Limit(100)
	db = db.Order("reviews_score DESC, FIELD(`type`,'game','dlc','demo','mod','video','movie','series','episode','application','tool','advertising') ASC, name ASC")
	db = db.Select([]string{"id", "name", "icon", "type", "platforms", "reviews_score"})
	db = db.Where("is_free = ?", "1")
	//db = db.Where("type = ?", "game")

	var apps []mysql.App

	db = db.Find(&apps)
	if db.Error != nil {
		logger.Error(db.Error)
		returnErrorTemplate(w, r, 500, "Can't connect to database")
		return
	}

	template := freeGamesTemplate{}
	template.Fill(w, r, "Free Games")
	template.Apps = apps

	returnTemplate(w, r, "free_games", template)
	return
}

type freeGamesTemplate struct {
	GlobalTemplate
	Apps []mysql.App
}
