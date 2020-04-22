package pages

import (
	"net/http"
	"sync"

	"github.com/gamedb/gamedb/cmd/webserver/helpers/session"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"go.mongodb.org/mongo-driver/bson"
)

func appsRandomHandler(w http.ResponseWriter, r *http.Request) {

	var t = appsRandomTemplate{}

	var player mongo.Player

	var filter = bson.D{
		{"name", bson.M{"$ne": ""}},
		{"type", "game"},
	}

	if session.IsLoggedIn(r) {

		ids := bson.A{}

		user, err := getUserFromSession(r)
		if err != nil {
			log.Err(err)
			returnErrorTemplate(w, r, errorTemplate{Code: 500})
			return
		}

		var steamID = user.GetSteamID()
		if steamID > 0 {

			player, err = mongo.GetPlayer(steamID)
			if err != nil {
				log.Err(err)
				returnErrorTemplate(w, r, errorTemplate{Code: 500})
				return
			}

			playerApps, err := mongo.GetPlayerApps(steamID, 0, 0, nil)
			if err != nil {
				log.Err(err)
				returnErrorTemplate(w, r, errorTemplate{Code: 500})
				return
			}

			for _, v := range playerApps {
				ids = append(ids, v.AppID)
			}

			filter = append(filter, bson.E{Key: "_id", Value: bson.M{"$in": ids}})
		}
	}

	var projection = bson.M{
		"name":               1,
		"type":               1,
		"background":         1,
		"movies":             1,
		"screenshots":        1,
		"achievements_count": 1,
		"tags":               1,
		"reviews_score":      1,
		"reviews_count":      1,
	}

	apps, err := mongo.GetRandomApps(1, filter, projection)
	if err != nil {
		log.Err(err)
		returnErrorTemplate(w, r, errorTemplate{Code: 500})
		return
	}
	if len(apps) == 0 {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "Couldn't find a game"})
		return
	}

	t.setBackground(apps[0], false, false)
	t.fill(w, r, "Random Steam Game", "Find a random Steam game")
	t.addAssetChosen()

	t.Apps = apps
	t.Player = player

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Tags, err = sql.GetTagsForSelect()
		log.Err(err, r)
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		if len(apps) > 0 {

			var err error
			t.AppTags, err = GetAppTags(apps[0])
			if err != nil {
				log.Err(err, r)
			}
		}
	}()

	wg.Wait()

	returnTemplate(w, r, "apps_random", t)
}

type appsRandomTemplate struct {
	GlobalTemplate
	Apps    []mongo.App
	Player  mongo.Player
	Tags    []sql.Tag
	AppTags []sql.Tag
}
