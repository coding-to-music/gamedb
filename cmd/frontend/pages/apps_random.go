package pages

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gamedb/gamedb/cmd/frontend/helpers/session"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/olivere/elastic/v7"
	"go.mongodb.org/mongo-driver/bson"
)

func appsRandomHandler(w http.ResponseWriter, r *http.Request) {

	var t = appsRandomTemplate{}
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {

		defer wg.Done()

		filters := []elastic.Query{
			elastic.NewTermsQuery("type", "game", ""),
			elastic.NewBoolQuery().MustNot(
				elastic.NewTermQuery("name.raw", ""),
			).MinimumNumberShouldMatch(1),
			elastic.NewBoolQuery().Should(
				elastic.NewRangeQuery("movies_count").From(1),
				elastic.NewRangeQuery("screenshots_count").From(1),
			).MinimumNumberShouldMatch(1),
		}

		var tag = r.URL.Query().Get("tag")
		if tag != "" {
			filters = append(filters, elastic.NewTermQuery("tags", tag))
		}

		var platform = r.URL.Query().Get("os")
		if helpers.SliceHasString(platform, []string{"windows", "macos", "linux"}) {
			filters = append(filters, elastic.NewTermQuery("platforms", platform))
		}

		var achievements = r.URL.Query().Get("achievements")
		if achievements != "" {
			filters = append(filters, elastic.NewRangeQuery("achievements_counts").From(1))
		}

		var popular = r.URL.Query().Get("popular")
		if popular != "" {
			filters = append(filters, elastic.NewRangeQuery("players").From(10))
		}

		var score = r.URL.Query().Get("score")
		if score != "" {
			filters = append(filters, elastic.NewRangeQuery("score").From(score))
		}

		var year = r.URL.Query().Get("year")
		if year != "" {
			now := time.Now()
			i, err := strconv.Atoi(year)
			if err == nil && i >= 1995 && i <= now.Year() {
				t := time.Date(i, 1, 1, 0, 0, 0, 0, now.Location())
				filters = append(filters, elastic.NewRangeQuery("release_date").From(t.Unix()))
			}
		}

		if session.IsLoggedIn(r) {

			var ids []interface{}

			player, err := getPlayerFromSession(r)
			if err != nil {
				log.ErrS(err)
				returnErrorTemplate(w, r, errorTemplate{Code: 500})
				return
			}

			if player.ID > 0 {

				t.Player = player

				var played = r.URL.Query().Get("played")
				if played != "" {

					playerApps, err := mongo.GetPlayerApps(0, 0, bson.D{{"player_id", player.ID}}, nil)
					if err != nil {
						log.ErrS(err)
						returnErrorTemplate(w, r, errorTemplate{Code: 500})
						return
					}

					for _, v := range playerApps {
						if played == "owned" || (played == "played" && v.AppTime > 0) || (played == "notplayed" && v.AppTime == 0) {
							ids = append(ids, v.AppID)
						}
					}

					filters = append(filters, elastic.NewTermsQuery("id", ids...))
				}
			}
		}

		app, count, err := elasticsearch.SearchAppsRandom(filters)
		if err != nil {
			err = helpers.IgnoreErrors(err, elasticsearch.ErrNoResult)
			if err != nil {
				log.ErrS(err)
			}
			return
		}

		t.App = app
		t.AppCount = count
		t.Price = app.Prices.Get(session.GetProductCC(r))
		t.setBackground(app, false, false)
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Tags, err = mongo.GetStatsForSelect(mongo.StatsTypeTags)
		if err != nil {
			log.ErrS(err)
		}
	}()

	wg.Wait()

	if t.App.ID > 0 {

		wg.Add(1)
		go func() {

			defer wg.Done()

			var err error
			t.AppTags, err = mongo.GetStatsByType(mongo.StatsTypeTags, t.App.Tags, t.App.ID)
			if err != nil {
				log.ErrS(err)
			}
		}()
	}

	wg.Wait()

	t.fill(w, r, "Random Steam Game", "Find a random Steam game")
	t.addAssetChosen()

	for i := time.Now().Year(); i >= 1995; i-- {
		t.Years = append(t.Years, i)
	}

	returnTemplate(w, r, "apps_random", t)
}

type appsRandomTemplate struct {
	globalTemplate
	App      elasticsearch.App
	AppCount int64
	Player   mongo.Player
	Tags     []mongo.Stat
	AppTags  []mongo.Stat
	Price    helpers.ProductPrice
	Years    []int
}
