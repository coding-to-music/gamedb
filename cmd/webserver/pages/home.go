package pages

import (
	"html/template"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/cmd/webserver/helpers/session"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/i18n"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
	"github.com/microcosm-cc/bluemonday"
	"go.mongodb.org/mongo-driver/bson"
)

func HomeRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/prices.json", homePricesHandler)
	r.Get("/sales/{sort}.json", homeSalesHandler)
	r.Get("/players/{sort}.json", homePlayersHandler)
	return r
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {

	t := homeTemplate{}
	t.setRandomBackground(true, true)
	t.fill(w, r, "Home", "Stats and information on the Steam Catalogue.")
	t.addAssetJSON2HTML()

	var wg sync.WaitGroup

	// Popular NEW games
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Games, err = mongo.PopularNewApps()
		if err != nil {
			log.Err(err, r)
		}
	}()

	// News
	wg.Add(1)
	go func() {

		defer wg.Done()

		apps, err := mongo.PopularApps()
		if err != nil {
			log.Err(err, r)
		}

		var appIDs []int
		var appIDmap = map[int]mongo.App{}
		for _, app := range apps {
			appIDs = append(appIDs, app.ID)
			appIDmap[app.ID] = app
		}

		news, err := mongo.GetArticlesByApps(appIDs, 20, time.Time{})
		if err != nil {
			log.Err(err, r)
		}

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
			})

			t.NewsID = v.ID
		}
	}()

	wg.Wait()

	var spotlights = []homeSpotlight{
		{"Discord Bot", "If you run a Discord chat server, we offer a bot to get player and game information!", "/chat-bot"},
		{"Experience Table", "Trying to level up and need to know how much XP you need?", "/experience"},
		{"Trending Groups", "Looking for trending groups to join?", "/groups?order=desc&sort=2"},
		{"Play with friends", "Find all the games you and your friends have in common and which ones are coop!", "/coop"},
		{"Game DB API", "Have a website and want to pull in information from Steam/Game DB?", "/api"},
		{"The most bans", "Curious who has been banned the most on all of Steam?", "/players?order=desc&sort=7#bans"},
	}

	t.Spotlight = spotlights[rand.Intn(len(spotlights))]

	//
	returnTemplate(w, r, "home", t)
}

type homeTemplate struct {
	GlobalTemplate
	Games     []mongo.App
	News      []homeNews
	NewsID    int64
	Players   []mongo.Player
	Spotlight homeSpotlight
}

type homeNews struct {
	Title    string
	Contents template.HTML
	Link     string
	Image    template.HTMLAttr
}

type homeSpotlight struct {
	Title string
	Text  template.HTML
	Link  string
}

func homePricesHandler(w http.ResponseWriter, r *http.Request) {

	var filter = bson.D{
		{Key: "prod_cc", Value: string(session.GetProductCC(r))},
		{Key: "app_id", Value: bson.M{"$gt": 0}},
		{Key: "difference", Value: bson.M{"$lt": 0}},
	}

	priceChanges, err := mongo.GetPrices(0, 10, filter)
	if err != nil {
		log.Err(err, r)
	}

	var prices []homePrice
	for _, price := range priceChanges {

		prices = append(prices, homePrice{
			Name:     price.Name,                                         // 0
			ID:       price.AppID,                                        // 1
			Link:     price.GetPath(),                                    // 2
			After:    i18n.FormatPrice(price.Currency, price.PriceAfter), // 3
			Discount: math.Round(price.DifferencePercent),                // 4
			Time:     price.CreatedAt.Unix(),                             // 5
			Avatar:   price.GetIcon(),                                    // 6
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

func homeSalesHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "sort")

	var sort string
	var order int

	switch id {
	case "top-rated":
		sort = "app_rating"
		order = -1
	case "ending-soon":
		sort = "offer_end"
		order = 1
	case "latest-found":
		sort = "badges_count"
		order = -1
	default:
		return
	}

	filter := bson.D{
		{Key: "app_type", Value: "game"},
		{Key: "sub_order", Value: 0},
		{Key: "offer_end", Value: bson.M{"$gt": time.Now()}},
	}

	sales, err := mongo.GetAllSales(0, 10, filter, bson.D{{Key: sort, Value: order}})
	if err != nil {
		log.Err(err)
	}

	var code = session.GetProductCC(r)

	var homeSales []homeSale
	for _, v := range sales {
		homeSales = append(homeSales, homeSale{
			ID:        v.AppID,
			Name:      v.AppName,
			Icon:      v.AppIcon,
			Type:      v.SaleType,
			Ends:      v.SaleEnd,
			Rating:    v.GetAppRating(),
			Price:     v.GetPriceString(code),
			Link:      helpers.GetAppPath(v.AppID, v.AppName),
			StoreLink: helpers.GetAppStoreLink(v.AppID),
		})
	}

	returnJSON(w, r, homeSales)
}

type homeSale struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Icon      string    `json:"icon"`
	Type      string    `json:"type"`
	Price     string    `json:"price"`
	Discount  int       `json:"discount"`
	Rating    string    `json:"rating"`
	Ends      time.Time `json:"ends"`
	Link      string    `json:"link"`
	StoreLink string    `json:"store_link"`
}

func homePlayersHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "sort")

	var sort string

	switch id {
	case "level":
		sort = "level"
	case "games":
		sort = "games_count"
	case "bans":
		sort = "bans_game"
	case "profile":
		sort = "friends_count"
	default:
		return
	}

	players, err := getPlayersForHome(sort)
	if err != nil {
		log.Err(err)
		return
	}

	var resp []homePlayer

	for k, player := range players {

		resp = append(resp, homePlayer{
			Name:     player.GetName(),
			Link:     player.GetPath(),
			Avatar:   player.GetAvatar(),
			Rank:     helpers.OrdinalComma(k + 1),
			Class:    helpers.GetPlayerAvatar2(player.Level),
			Level:    humanize.Comma(int64(player.Level)),
			Badges:   humanize.Comma(int64(player.BadgesCount)),
			Games:    humanize.Comma(int64(player.GamesCount)),
			Playtime: helpers.GetTimeLong(player.PlayTime, 2),
			GameBans: humanize.Comma(int64(player.NumberOfGameBans)),
			VACBans:  humanize.Comma(int64(player.NumberOfVACBans)),
			Friends:  humanize.Comma(int64(player.FriendsCount)),
			Comments: humanize.Comma(int64(player.CommentsCount)),
		})
	}

	returnJSON(w, r, resp)
}

type homePlayer struct {
	Rank     string `json:"rank"`
	Name     string `json:"name"`
	Link     string `json:"link"`
	Avatar   string `json:"avatar"`
	Class    string `json:"class"`
	Level    string `json:"level"`
	Badges   string `json:"badges"`
	Games    string `json:"games"`
	Playtime string `json:"playtime"`
	GameBans string `json:"game_bans"`
	VACBans  string `json:"vac_bans"`
	Friends  string `json:"friends"`
	Comments string `json:"comments"`
}

func getPlayersForHome(sort string) (players []mongo.Player, err error) {

	var item = memcache.MemcacheHomePlayers(sort)

	err = memcache.GetClient().GetSetInterface(item.Key, item.Expiration, &players, func() (interface{}, error) {

		projection := bson.M{
			"_id":            1,
			"persona_name":   1,
			"avatar":         1,
			"level":          1,
			"badges_count":   1,
			"games_count":    1,
			"play_time":      1,
			"bans_game":      1,
			"bans_cav":       1,
			"friends_count":  1,
			"comments_count": 1,
		}

		return mongo.GetPlayers(0, 10, bson.D{{Key: sort, Value: -1}}, bson.D{{Key: sort, Value: bson.M{"$gt": 0}}}, projection)
	})

	return players, err
}
