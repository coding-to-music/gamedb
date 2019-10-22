package pages

import (
	"html/template"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
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

	var wg sync.WaitGroup

	// Popular NEW games
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
			contents = helpers.TruncateString(contents, 300, "...")
			contents = strings.TrimSpace(contents)

			t.News = append(t.News, homeNews{
				Title:    v.Title,
				Contents: template.HTML(contents),
				Link:     "/news#" + strconv.FormatInt(v.ID, 10),
				Image:    template.HTMLAttr(appIDmap[v.AppID].GetHeaderImage()),
				Image2:   template.HTMLAttr(appIDmap[v.AppID].GetHeaderImage()),
			})

			t.NewsID = v.ID
		}
	}()

	wg.Wait()

	//
	returnTemplate(w, r, "home", t)
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
	Image2   template.HTMLAttr
}

func homePricesHandler(w http.ResponseWriter, r *http.Request) {

	var filter = mongo.D{
		{"prod_cc", string(helpers.GetProductCC(r))},
		{"app_id", bson.M{"$gt": 0}},
		{"difference", bson.M{"$lt": 0}},
	}

	priceChanges, err := mongo.GetPrices(0, 15, filter)
	log.Err(err, r)

	var prices []homePrice
	for _, price := range priceChanges {

		prices = append(prices, homePrice{
			Name:     helpers.InsertNewLines(price.Name),                    // 0
			ID:       price.AppID,                                           // 1
			Link:     price.GetPath(),                                       // 2
			After:    helpers.FormatPrice(price.Currency, price.PriceAfter), // 3
			Discount: math.Round(price.DifferencePercent),                   // 4
			Time:     price.CreatedAt.Unix(),                                // 5
			Avatar:   price.GetIcon(),                                       // 6
		})
	}

	returnJSON(w, r, prices)
}

type homePrice struct {
	Name     string  `json:"name"`
	ID       int     `json:"id"`
	Link     string  `json:"link"`
	After    string  `json:"after"`
	Discount float64 `json:"discount"`
	Time     int64   `json:"time"`
	Avatar   string  `json:"avatar"`
}

func homePlayersHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "sort")

	if !helpers.SliceHasString([]string{"level", "games", "badges", "time", "comments", "friends"}, id) {
		return
	}

	var sort string

	switch id {
	case "level":
		sort = "level"
	case "games":
		sort = "games_count"
	case "badges":
		sort = "badges_count"
	case "time":
		sort = "play_time"
	case "friends":
		sort = "friends_count"
	case "comments":
		sort = "comments_count"
	default:
		return
	}

	players, err := getPlayersForHome(sort)
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
			homePlayer.setRank(player.GetRank(mongo.RankKeyLevel, mongo.RankCountryAll))
			homePlayer.Value = humanize.Comma(int64(player.Level))
			homePlayer.Class = helpers.GetPlayerAvatar2(player.Level)
		case "games":
			homePlayer.setRank(player.GetRank(mongo.RankKeyGames, mongo.RankCountryAll))
			homePlayer.Value = humanize.Comma(int64(player.GamesCount))
		case "badges":
			homePlayer.setRank(player.GetRank(mongo.RankKeyBadges, mongo.RankCountryAll))
			homePlayer.Value = humanize.Comma(int64(player.BadgesCount))
		case "time":
			homePlayer.setRank(player.GetRank(mongo.RankKeyPlaytime, mongo.RankCountryAll))
			homePlayer.Value = helpers.GetTimeLong(player.PlayTime, 2)
		case "friends":
			homePlayer.setRank(player.GetRank(mongo.RankKeyFriends, mongo.RankCountryAll))
			homePlayer.Value = humanize.Comma(int64(player.FriendsCount))
		case "comments":
			homePlayer.setRank(player.GetRank(mongo.RankKeyComments, mongo.RankCountryAll))
			homePlayer.Value = humanize.Comma(int64(player.CommentsCount))
		}

		resp = append(resp, homePlayer)
	}

	returnJSON(w, r, resp)
}

type homePlayer struct {
	Rank   string `json:"rank"`
	Name   string `json:"name"`
	Value  string `json:"value"`
	Link   string `json:"link"`
	Avatar string `json:"avatar"`
	Class  string `json:"class"`
}

func (hp *homePlayer) setRank(i int, b bool) {
	if b {
		hp.Rank = helpers.OrdinalComma(i)
	} else {
		hp.Rank = "-"
	}
}

func getPlayersForHome(sort string) (players []mongo.Player, err error) {

	var item = helpers.MemcacheHomePlayers(sort)

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &players, func() (interface{}, error) {

		projection := mongo.M{
			"_id":          1,
			"persona_name": 1,
			"avatar":       1,
			"ranks":        1,
			sort:           1,
		}

		return mongo.GetPlayers(0, 15, mongo.D{{sort, -1}}, mongo.M{sort: mongo.M{"$gt": 0}}, projection)
	})

	return players, err
}
