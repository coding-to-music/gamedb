package pages

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"go.mongodb.org/mongo-driver/bson"
)

func appsRandomHandler(w http.ResponseWriter, r *http.Request) {

	var player mongo.Player

	var filter = bson.D{
		{"$or", bson.A{
			bson.M{"type": "game"},
			bson.M{"type": ""},
		}},
		{Key: "$or", Value: bson.A{
			bson.M{"movies": bson.D{{Key: "$exists", Value: true}, {Key: "$ne", Value: bson.A{}}, {Key: "$ne", Value: nil}}},
			bson.M{"screenshots": bson.D{{Key: "$exists", Value: true}, {Key: "$ne", Value: bson.A{}}, {Key: "$ne", Value: nil}}},
		}},
		{"name", bson.M{"$ne": ""}},
	}

	var tag = r.URL.Query().Get("tag")
	if tag != "" {
		i, err := strconv.Atoi(tag)
		if err == nil && i > 0 {
			filter = append(filter, bson.E{Key: "tags", Value: i})
		}
	}

	var platform = r.URL.Query().Get("os")
	if helpers.SliceHasString(platform, []string{"windows", "macos", "linux"}) {
		filter = append(filter, bson.E{Key: "platforms", Value: platform})
	}

	var achievements = r.URL.Query().Get("achievements")
	if achievements != "" {
		filter = append(filter, bson.E{Key: "achievements_count", Value: bson.M{"$gt": 0}})
	}

	var popular = r.URL.Query().Get("popular")
	if popular != "" {
		filter = append(filter, bson.E{Key: "player_peak_alltime", Value: bson.M{"$gte": 10}})
	}

	var score = r.URL.Query().Get("score")
	if score != "" {
		i, err := strconv.Atoi(score)
		if err == nil && i > 0 {
			filter = append(filter, bson.E{Key: "reviews_score", Value: bson.M{"$gte": i}})
		}
	}

	var year = r.URL.Query().Get("year")
	if year != "" {
		now := time.Now()
		i, err := strconv.Atoi(year)
		if err == nil && i >= 1995 && i <= now.Year() {
			t := time.Date(i, 1, 1, 0, 0, 0, 0, now.Location())
			filter = append(filter, bson.E{Key: "release_date_unix", Value: bson.M{"$gte": t.Unix()}})
		}
	}

	var t = appsRandomTemplate{}

	if session.IsLoggedIn(r) {

		ids := bson.A{}

		user, err := getUserFromSession(r)
		if err != nil {
			log.Err(err, r)
			returnErrorTemplate(w, r, errorTemplate{Code: 500})
			return
		}

		var steamID = user.GetSteamID()
		if steamID > 0 {

			player, err = mongo.GetPlayer(steamID)
			if err != nil {
				log.Err(err, r)
				returnErrorTemplate(w, r, errorTemplate{Code: 500})
				return
			}

			t.Player = player

			var played = r.URL.Query().Get("played")
			if played != "" {

				playerApps, err := mongo.GetPlayerApps(0, 0, bson.D{{"player_id", steamID}}, nil)
				if err != nil {
					log.Err(err, r)
					returnErrorTemplate(w, r, errorTemplate{Code: 500})
					return
				}

				for _, v := range playerApps {
					if played == "owned" || (played == "played" && v.AppTime > 0) || (played == "notplayed" && v.AppTime == 0) {
						ids = append(ids, v.AppID)
					}
				}

				filter = append(filter, bson.E{Key: "_id", Value: bson.M{"$in": ids}})
			}
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
		"prices":             1,
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Apps, err = mongo.GetRandomApps(1, filter, projection)
		if err != nil {
			log.Err(err, r)
		}

		if len(t.Apps) > 0 {
			t.setBackground(t.Apps[0], false, false)
			t.Price = t.Apps[0].Prices.Get(session.GetProductCC(r))
		}
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.AppCount, err = mongo.CountDocuments(mongo.CollectionApps, filter, 0)
		if err != nil {
			log.Err(err, r)
		}
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Tags, err = mysql.GetTagsForSelect()
		log.Err(err, r)
	}()

	wg.Wait()

	t.fill(w, r, "Random Steam Game", "Find a random Steam game")
	t.addAssetChosen()

	for i := time.Now().Year(); i >= 1995; i-- {
		t.Years = append(t.Years, i)
	}

	if len(t.Apps) > 0 {

		wg.Add(1)
		go func() {

			defer wg.Done()

			var err error
			t.AppTags, err = GetAppTags(t.Apps[0])
			if err != nil {
				log.Err(err, r)
			}
		}()
	}

	wg.Wait()

	returnTemplate(w, r, "apps_random", t)
}

type appsRandomTemplate struct {
	GlobalTemplate
	Apps     []mongo.App
	AppCount int64
	Player   mongo.Player
	Tags     []mysql.Tag
	AppTags  []mysql.Tag
	Price    helpers.ProductPrice
	Years    []int
}
