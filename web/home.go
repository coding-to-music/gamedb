package web

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gamedb/website/config"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/mongo"
	"github.com/gamedb/website/session"
	"github.com/gamedb/website/sql"
	"github.com/go-chi/chi"
	"github.com/microcosm-cc/bluemonday"
	"go.mongodb.org/mongo-driver/bson"
)

func homeRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/prices.json", homePricesHandler)
	r.Get("/{sort}/players.json", homePlayersHandler)
	return r
}

func homeHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	setCacheHeaders(w, time.Hour)

	t := homeTemplate{}
	t.fill(w, r, "Home", "Stats and information on the Steam Catalogue.")
	t.addAssetJSON2HTML()

	var wg sync.WaitGroup

	// Popular games
	wg.Add(1)
	go func() {

		defer wg.Done()

		gorm, err := sql.GetMySQLClient()
		if err != nil {
			log.Err(err)
			return
		}

		gorm = gorm.Select([]string{"id", "name", "image_header"})
		gorm = gorm.Where("type = ?", "game")
		gorm = gorm.Where("release_date_unix > ?", time.Now().AddDate(0, 0, -config.Config.NewReleaseDays).Unix())
		gorm = gorm.Order("player_peak_week desc")
		gorm = gorm.Limit(20)
		gorm = gorm.Find(&t.Games)

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

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	setCacheHeaders(w, time.Minute)

	var filter = mongo.D{
		{"currency", string(session.GetCountryCode(r))},
		{"app_id", bson.M{"$gt": 0}},
		{"difference", bson.M{"$lt": 0}},
	}

	priceChanges, err := mongo.GetPrices(0, 15, filter)
	log.Err(err, r)

	locale, err := helpers.GetLocaleFromCountry(session.GetCountryCode(r))
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

	b, err := json.Marshal(prices)
	if err != nil {
		log.Err(err)
		return
	}

	err = returnJSON(w, r, b)
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

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	setCacheHeaders(w, time.Hour*6)

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

	players, err := mongo.GetPlayers(0, 10, mongo.D{{sort, 1}}, mongo.M{sort: mongo.M{"$gt": 0}}, projection)
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

	b, err := json.Marshal(resp)
	if err != nil {
		log.Err(err)
		return
	}

	err = returnJSON(w, r, b)
	log.Err(err)
}

type homePlayer struct {
	Rank   string `json:"rank"`
	Name   string `json:"name"`
	Value  int    `json:"value"`
	Link   string `json:"link"`
	Avatar string `json:"avatar"`
}
