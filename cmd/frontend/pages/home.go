package pages

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/session"
	twitterHelper "github.com/gamedb/gamedb/cmd/frontend/helpers/twitter"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/go-chi/chi"
	"github.com/mborgerson/GoTruncateHtml/truncatehtml"
	"github.com/microcosm-cc/bluemonday"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
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

var (
	regexpAppID = regexp.MustCompile(`/app/([0-9]+)`)
	regexpSubID = regexp.MustCompile(`/sub/([0-9]+)`)
)

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
			log.ErrS(err)
		}

		if len(t.NewGames) > 10 {
			t.NewGames = t.NewGames[0:10]
		}
	}()

	// Top games
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.TopGames, err = mongo.PopularApps()
		if err != nil {
			log.ErrS(err)
		}

		if len(t.TopGames) > 10 {
			t.TopGames = t.TopGames[0:10]
		}
	}()

	// Top sellers
	wg.Add(1)
	go func() {

		defer wg.Done()

		var topSellers []homeTopSellerTemplate

		callback := func() (interface{}, error) {

			b, _, err := helpers.Get("https://store.steampowered.com/feeds/weeklytopsellers.xml", 0, nil)
			if err != nil {
				return b, err
			}

			vdf := RDF{}
			err = xml.Unmarshal(b, &vdf)
			if err != nil {
				return b, err
			}

			for _, v := range vdf.Channel.Seq.Li {

				matches := regexpAppID.FindStringSubmatch(v.Resource)
				if len(matches) == 2 {
					i, err := strconv.Atoi(matches[1])
					if err == nil {

						app, err := mongo.GetApp(i)
						if err != nil {
							log.ErrS(err, zap.Int("app", i))
							continue
						}

						topSellers = append(topSellers, homeTopSellerTemplate{
							ID:    app.ID,
							Path:  app.GetPath(),
							Name:  app.GetName(),
							Image: app.GetHeaderImage(),
							Type:  helpers.ProductTypeApp,
						})
					}
				}

				matches = regexpSubID.FindStringSubmatch(v.Resource)
				if len(matches) == 2 {
					i, err := strconv.Atoi(matches[1])
					if err == nil {

						sub, err := mongo.GetPackage(i)
						if err != nil {
							log.ErrS(err, zap.Int("sub", i))
							continue
						}

						// Force absolute for images.weserv.nl
						image := sub.ImagePage
						if image == "" {
							image = helpers.DefaultAppIcon
						}
						if strings.HasPrefix(image, "/") {
							image = "http://gamedb.online" + image // Not using domain from config to make local work
						}

						topSellers = append(topSellers, homeTopSellerTemplate{
							ID:    sub.ID,
							Path:  sub.GetPath(),
							Name:  sub.GetName(),
							Image: image,
							Type:  helpers.ProductTypePackage,
						})
					}
				}
			}

			return topSellers, nil
		}

		item := memcache.HomeTopSellers
		err := memcache.GetSetInterface(item.Key, item.Expiration, &topSellers, callback)
		if err != nil {
			log.ErrS(err)
			return
		}

		t.TopSellers = topSellers
	}()

	wg.Wait()

	t.ConstApp = helpers.ProductTypeApp
	t.ConstPackage = helpers.ProductTypePackage

	//
	returnTemplate(w, r, "home", t)
}

type homeTemplate struct {
	globalTemplate
	TopGames     []mongo.App
	NewGames     []mongo.App
	Players      []mongo.Player
	TopSellers   []homeTopSellerTemplate
	ConstApp     helpers.ProductType
	ConstPackage helpers.ProductType
}

type homeTopSellerTemplate struct {
	ID    int
	Path  string
	Name  string
	Image string
	Type  helpers.ProductType
}

type RDF struct {
	Channel struct {
		Seq struct {
			Text string `xml:",chardata"`
			Li   []struct {
				Resource string `xml:"resource,attr"`
			} `xml:"li"`
		} `xml:"Seq"`
	} `xml:"channel"`
}

var htmlPolicy = bluemonday.
	NewPolicy().
	AllowElements("br").
	AllowAttrs("data-lazy").Globally()

func homeNewsHandler(w http.ResponseWriter, r *http.Request) {

	t := homeNewsTemplate{}
	t.fill(w, r, "", "")

	apps, err := mongo.PopularApps()
	if err != nil {
		log.ErrS(err)
	}

	var appIDs []int
	var appIDmap = map[int]mongo.App{}
	for _, app := range apps {
		appIDs = append(appIDs, app.ID)
		appIDmap[app.ID] = app
	}

	news, err := mongo.GetArticlesByAppIDs(appIDs, 20, time.Time{})
	if err != nil {
		log.ErrS(err)
	}

	for _, v := range news {

		contents := string(v.GetBody())
		contents = htmlPolicy.Sanitize(contents)
		contents = helpers.RegexSpacesStartEnd.ReplaceAllString(contents, "")

		b, err := truncatehtml.TruncateHtml([]byte(contents), 200, "...")
		if err == nil {
			contents = string(b)
		}

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

	var item = memcache.HomeTweets
	var ret []homeTweet

	callback := func() (interface{}, error) {

		t := true
		f := false

		params := &twitter.UserTimelineParams{
			ScreenName:      "gamedb_online",
			Count:           4,
			ExcludeReplies:  &t,
			IncludeRetweets: &f,
		}

		tweets, resp, err := twitterHelper.GetTwitter().Timelines.UserTimeline(params)
		if err != nil {
			return nil, err
		}

		defer helpers.Close(resp.Body)

		for _, v := range tweets {
			ret = append(ret, homeTweet{
				ScreenName: v.User.ScreenName,
				Name:       v.User.Name,
				Avatar:     v.User.ProfileImageURLHttps,
				Text:       v.Text,
				Link:       fmt.Sprintf("https://twitter.com/%s/status/%s", v.User.ScreenName, v.IDStr),
			})
		}

		return ret, nil
	}

	err := memcache.GetSetInterface(item.Key, item.Expiration, &ret, callback)
	if err != nil {
		log.ErrS(err)
		return
	}

	returnJSON(w, r, ret)
}

type homeTweet struct {
	ScreenName string `json:"screen_name"`
	Name       string `json:"name"`
	Avatar     string `json:"avatar"`
	Text       string `json:"text"`
	Link       string `json:"link"`
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
		log.ErrS(err)
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
		log.ErrS(err)
		return
	}

	var resp []queue.PlayerPayload
	for _, player := range players {
		resp = append(resp, queue.PlayerPayload{
			ID:            strconv.FormatInt(player.ID, 10),
			Name:          player.GetName(),
			Avatar:        player.GetAvatar(),
			Link:          player.GetPath(),
			CommunityLink: player.CommunityLink(),
			UpdatedAt:     player.UpdatedAt.Unix(),
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
		log.ErrS(err)
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
