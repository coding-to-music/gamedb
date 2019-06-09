package pages

import (
	"html/template"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/go-chi/chi"
	"github.com/microcosm-cc/bluemonday"
	"go.mongodb.org/mongo-driver/bson"
)

func HomeRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/prices.json", homePricesHandler)
	r.Get("/{sort}/players.json", homePlayersHandler)
	return r
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {

	t := homeTemplate{}
	t.fill(w, r, "Home", "Stats and information on the Steam Catalogue.")
	t.addAssetJSON2HTML()
	t.setFlashes(w, r, true)
	t.Background = func() string {
		apps := []int{
			10, 70, 220, 240, 400, 400, 620, 8800, 8930, 9420, 22380, 48700, 49520, 105600, 113200, 206190, 213670, 213670, 219740, 223470,
			268910, 280220, 284810, 292030, 294100, 294100, 311190, 312530, 324160, 367520, 385250, 386940, 411960, 413150, 415000, 415000,
			427520, 460950, 525510, 591460, 597220, 597220, 611760, 635940, 704470, 816490, 843380, 883710, 910630, 942970, 1036580, 1046030,
		}
		return "https://steamcdn-a.akamaihd.net/steam/apps/" + strconv.Itoa(apps[rand.Intn(len(apps))]) + "/page_bg_generated_v6b.jpg"
	}()

	var wg sync.WaitGroup

	// Popular games
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Games, err = sql.PopularNewApps()
		log.Err(err, r)
	}()

	// News
	wg.Add(1)
	go func() {

		defer wg.Done()

		apps, err := sql.PopularApps()
		log.Err(err, r)

		var appIDs []int
		var appIDmap = map[int]sql.App{}
		for _, app := range apps {
			appIDs = append(appIDs, app.ID)
			appIDmap[app.ID] = app
		}

		news, err := mongo.GetArticlesByApps(appIDs, 20, time.Time{})
		log.Err(err, r)

		p := bluemonday.StrictPolicy() // Strip all tags

		for _, v := range news {

			contents := string(helpers.RenderHTMLAndBBCode(v.Contents))
			contents = p.Sanitize(contents)
			contents = helpers.TruncateString(contents, 300)
			contents = strings.TrimSpace(contents)

			t.News = append(t.News, homeNews{
				Title:    v.Title,
				Contents: template.HTML(contents),
				Link:     "/news#" + strconv.FormatInt(v.ID, 10),
				Image:    template.HTMLAttr(appIDmap[v.AppID].GetHeaderImage()),
			})

			t.NewsID = v.ID
		}
	}()

	wg.Wait()

	//
	err := returnTemplate(w, r, "home", t)
	log.Err(err, r)
}

type homeTemplate struct {
	GlobalTemplate
	Games   []sql.App
	News    []homeNews
	NewsID  int64
	Players []mongo.Player
}

type homeNews struct {
	Title    string
	Contents template.HTML
	Link     string
	Image    template.HTMLAttr
}

func homePricesHandler(w http.ResponseWriter, r *http.Request) {

	var filter = mongo.D{
		{"currency", string(helpers.GetCountryCode(r))},
		{"app_id", bson.M{"$gt": 0}},
		{"difference", bson.M{"$lt": 0}},
	}

	priceChanges, err := mongo.GetPrices(0, 15, filter)
	log.Err(err, r)

	locale, err := helpers.GetLocaleFromCountry(helpers.GetCountryCode(r))
	log.Err(err)

	var prices []homePrice

	for _, v := range priceChanges {

		prices = append(prices, homePrice{
			Name:   v.Name,
			ID:     v.AppID,
			Link:   v.GetPath(),
			Before: locale.Format(v.PriceBefore),
			After:  locale.Format(v.PriceAfter),
			Time:   v.CreatedAt.Unix(),
			Avatar: v.GetIcon(),
		})
	}

	err = returnJSON(w, r, prices)
	log.Err(err)
}

type homePrice struct {
	Name   string `json:"name"`
	ID     int    `json:"id"`
	Link   string `json:"link"`
	Before string `json:"before"`
	After  string `json:"after"`
	Time   int64  `json:"time"`
	Avatar string `json:"avatar"`
}

func homePlayersHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "sort")

	if !helpers.SliceHasString([]string{"level", "games", "badges", "time"}, id) {
		return
	}

	var sort string
	var value string

	switch id {
	case "level":
		sort = "level_rank"
		value = "level"
	case "games":
		sort = "games_rank"
		value = "games_count"
	case "badges":
		sort = "badges_rank"
		value = "badges_count"
	case "time":
		sort = "play_time_rank"
		value = "play_time"
	}

	projection := mongo.M{
		"_id":          1,
		"persona_name": 1,
		"avatar":       1,
		sort:           1,
		value:          1,
	}

	players, err := mongo.GetPlayers(0, 10, mongo.D{{sort, 1}}, mongo.M{sort: mongo.M{"$gt": 0}}, projection, nil)
	if err != nil {
		log.Err(err)
		return
	}

	var resp []homePlayer

	for _, player := range players {

		homePlayer := homePlayer{
			Name:   player.GetName(),
			Link:   player.GetPath(),
			Avatar: player.GetAvatar(),
		}

		switch id {
		case "level":
			homePlayer.Rank = player.GetLevelRank()
			homePlayer.Value = player.Level
		case "games":
			homePlayer.Rank = player.GetGamesRank()
			homePlayer.Value = player.GamesCount
		case "badges":
			homePlayer.Rank = player.GetBadgesRank()
			homePlayer.Value = player.BadgesCount
		case "time":
			homePlayer.Rank = player.GetPlaytimeRank()
			homePlayer.Value = player.PlayTime
		}

		resp = append(resp, homePlayer)
	}

	err = returnJSON(w, r, resp)
	log.Err(err)
}

type homePlayer struct {
	Rank   string `json:"rank"`
	Name   string `json:"name"`
	Value  int    `json:"value"`
	Link   string `json:"link"`
	Avatar string `json:"avatar"`
}
