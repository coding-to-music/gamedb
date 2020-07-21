package pages

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	twitter2 "github.com/dghubble/go-twitter/twitter"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/twitter"
	"github.com/go-chi/chi"
	"github.com/microcosm-cc/bluemonday"
	"go.mongodb.org/mongo-driver/bson"
)

func HomeRouter() http.Handler {
	r := chi.NewRouter()
	// r.Get("/sales/{sort}.json", homeSalesHandler)
	r.Get("/players/{sort}.json", homePlayersHandler)
	r.Get("/updated-players.json", homeUpdatedPlayersHandler)
	r.Get("/tweets.json", homeTweetsHandler)
	r.Get("/news.html", homeNewsHandler)
	return r
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {

	t := homeTemplate{}
	t.setRandomBackground(true, true)
	t.fill(w, r, "Home", "Stats and information on the Steam Catalogue.")
	t.addAssetJSON2HTML()

	var wg sync.WaitGroup

	// New games
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.NewGames, err = mongo.PopularNewApps()
		if err != nil {
			log.Err(err, r)
		}
	}()

	// Top games
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.TopGames, err = mongo.PopularApps()
		if err != nil {
			log.Err(err, r)
		}

		if len(t.TopGames) > 12 {
			t.TopGames = t.TopGames[0:12]
		}
	}()

	wg.Wait()

	//
	returnTemplate(w, r, "home", t)
}

type homeTemplate struct {
	globalTemplate
	TopGames []mongo.App
	NewGames []mongo.App
	Players  []mongo.Player
}

func homeNewsHandler(w http.ResponseWriter, r *http.Request) {

	t := homeNewsTemplate{}
	t.fill(w, r, "", "")

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

	news, err := mongo.GetArticlesByAppIDs(appIDs, 20, time.Time{})
	if err != nil {
		log.Err(err, r)
	}

	p := bluemonday.StrictPolicy() // Strip all tags

	for _, v := range news {

		contents := string(v.GetBody())
		contents = p.Sanitize(contents)
		contents = helpers.TruncateString(contents, 300, "...")
		contents = strings.TrimSpace(contents)

		t.News = append(t.News, homeNewsItemTemplate{
			Title:    v.Title,
			Contents: template.HTML(contents),
			Link:     "/news#" + strconv.FormatInt(v.ID, 10),
			Image:    template.HTMLAttr(appIDmap[v.AppID].GetHeaderImage()),
		})

		t.NewsID = v.ID
	}

	returnTemplate(w, r, "home_news", t)
}

type homeNewsTemplate struct {
	globalTemplate
	News   []homeNewsItemTemplate
	NewsID int64
}

type homeNewsItemTemplate struct {
	Title    string
	Contents template.HTML
	Link     string
	Image    template.HTMLAttr
}

func homeTweetsHandler(w http.ResponseWriter, r *http.Request) {

	t := true
	f := false

	tweets, resp, err := twitter.GetTwitter().Timelines.UserTimeline(&twitter2.UserTimelineParams{
		ScreenName:      "gamedbonline",
		Count:           10,
		ExcludeReplies:  &t,
		IncludeRetweets: &f,
	})

	log.Err(err)
	log.Err(resp.Body.Close())

	log.Info(tweets)

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
		log.Err(err, r)
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

func homeUpdatedPlayersHandler(w http.ResponseWriter, r *http.Request) {

	var projection = bson.M{
		"_id":          1,
		"persona_name": 1,
		"avatar":       1,
		"updated_at":   1,
	}

	players, err := mongo.GetPlayers(0, 10, bson.D{{"updated_at", -1}}, nil, projection)
	if err != nil {
		log.Err(err, r)
		return
	}

	var resp []queue.PlayerPayload
	for _, player := range players {
		resp = append(resp, queue.PlayerPayload{
			ID:        strconv.FormatInt(player.ID, 10),
			Name:      player.GetName(),
			Avatar:    player.GetAvatar(),
			Link:      player.GetPath(),
			UpdatedAt: player.UpdatedAt.Unix(),
		})
	}

	returnJSON(w, r, resp)
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
		log.Err(err, r)
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

	err = memcache.GetSetInterface(item.Key, item.Expiration, &players, func() (interface{}, error) {

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
