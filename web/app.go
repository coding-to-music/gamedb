package web

import (
	"encoding/json"
	"html/template"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/logging"
	"github.com/gamedb/website/session"
	"github.com/go-chi/chi"
	"github.com/grokify/html-strip-tags-go"
)

func AppHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	if id == "" {
		returnErrorTemplate(w, r, 400, "Invalid App ID")
		return
	}

	idx, err := strconv.Atoi(id)
	if err != nil {
		returnErrorTemplate(w, r, 400, "Invalid App ID: "+id)
		return
	}

	if !db.IsValidAppID(idx) {
		returnErrorTemplate(w, r, 400, "Invalid App ID: "+id)
		return
	}

	// Get app
	app, err := db.GetApp(idx)
	if err != nil {

		if err == db.ErrCantFindApp {
			returnErrorTemplate(w, r, 404, "Sorry but we can not find this app.")
			return
		}

		returnErrorTemplate(w, r, 500, err.Error())
		return
	}

	// Redirect to correct slug
	if r.URL.Path != app.GetPath() {
		http.Redirect(w, r, app.GetPath(), 302)
		return
	}

	// Update news, reviews etc
	errs := app.UpdateFromRequest(r.UserAgent())
	for _, v := range errs {
		logging.Error(v)
	}

	//
	var wg sync.WaitGroup

	// todo, dont call steam here!
	var achievements []appAchievementTemplate
	//wg.Add(1)
	//go func() {
	//
	//	// Get achievements
	//	achievementsResp, _, err := helpers.GetSteam().GetGlobalAchievementPercentagesForApp(app.ID)
	//	if err != nil {
	//
	//		logging.Error(err)
	//
	//	} else {
	//
	//		achievementsMap := make(map[string]float64)
	//		for _, v := range achievementsResp.GlobalAchievementPercentage {
	//			achievementsMap[v.Name] = v.Percent
	//		}
	//
	//		// Get schema
	//		schema, _, err := helpers.GetSteam().GetSchemaForGame(app.ID)
	//		if err != nil {
	//
	//			logging.Error(err)
	//
	//		} else {
	//
	//			// Make template struct
	//			for _, v := range schema.AvailableGameStats.Achievements {
	//				achievements = append(achievements, appAchievementTemplate{
	//					v.Icon,
	//					v.DisplayName,
	//					v.Description,
	//					achievementsMap[v.Name],
	//				})
	//			}
	//		}
	//	}
	//
	//	wg.Done()
	//}()

	var tags []db.Tag
	wg.Add(1)
	go func() {

		// Get tags
		tags, err = app.GetTags()
		if err != nil {
			logging.Error(err)
		}

		wg.Done()
	}()

	var pricesString string
	var pricesCount int
	wg.Add(1)
	go func() {

		var code = session.GetCountryCode(r)

		// Get prices
		pricesResp, err := db.GetProductPrices(app.ID, db.ProductTypeApp, code)
		if err != nil {

			logging.Error(err)

		} else {

			pricesCount = len(pricesResp)

			var prices [][]float64

			for _, v := range pricesResp {

				prices = append(prices, []float64{float64(v.CreatedAt.Unix()), float64(v.PriceAfter) / 100})
			}

			// Add current price
			price := app.GetPrice(code)

			prices = append(prices, []float64{float64(time.Now().Unix()), float64(price.Final) / 100})

			// Make into a JSON string
			pricesBytes, err := json.Marshal(prices)
			if err != nil {

				logging.Error(err)

			} else {

				pricesString = string(pricesBytes)

			}
		}

		wg.Done()
	}()

	var news []appArticleTemplate
	wg.Add(1)
	go func() {

		// Get news
		newsResp, err := db.GetAppArticles(idx)
		if err != nil {

			logging.Error(err)

		} else {

			// todo, use a different bbcode library that works for app 418460 & 218620
			// todo, add http to links here instead of JS
			//var regex = regexp.MustCompile(`href="(?!http)(.*)"`)
			//var conv bbConvert.HTMLConverter
			//conv.ImplementDefaults()

			for _, v := range newsResp {

				// Fix broken links
				//v.Contents = regex.ReplaceAllString(v.Contents, `$1http://$2`)

				// Convert BBCdoe to HTML
				//v.Contents = conv.Convert(v.Contents)

				news = append(news, appArticleTemplate{
					ID:       v.ArticleID,
					Title:    v.Title,
					Contents: template.HTML(strip.HTMLEscapeString(v.Contents)),
					Author:   v.Author,
				})
			}

		}

		wg.Done()
	}()

	var packages []db.Package
	wg.Add(1)
	go func() {

		// Get packages
		packages, err = db.GetPackagesAppIsIn(app.ID)
		if err != nil {
			logging.Error(err)
		}

		wg.Done()
	}()

	var dlc []db.App
	wg.Add(1)
	go func() {

		// Get DLC
		dlc, err = db.GetDLC(app, []string{"id", "name"})
		if err != nil {
			logging.Error(err)
		}

		wg.Done()
	}()

	// Get reviews
	var reviews []appReviewTemplate
	var reviewsCount steam.ReviewsSummaryResponse
	wg.Add(1)
	go func() {

		reviewsResponse, err := app.GetReviews()
		if err != nil {

			logging.Error(err)

		} else {

			reviewsCount = reviewsResponse.QuerySummary

			// Make slice of playerIDs
			var playerIDs []int64
			for _, v := range reviewsResponse.Reviews {
				playerIDs = append(playerIDs, v.Author.SteamID)
			}

			players, err := db.GetPlayersByIDs(playerIDs)
			if err != nil {

				logging.Error(err)

			} else {

				// Make map of players
				var playersMap = map[int64]db.Player{}
				for _, v := range players {
					playersMap[v.PlayerID] = v
				}

				// Make template slice
				for _, v := range reviewsResponse.Reviews {

					var player db.Player
					if val, ok := playersMap[v.Author.SteamID]; ok {
						player = val
					} else {
						player = db.Player{}
						player.PlayerID = v.Author.SteamID
						player.PersonaName = "Unknown"
					}

					// Remove extra new lines
					regex := regexp.MustCompile("[\n]{3,}") // After comma
					v.Review = regex.ReplaceAllString(v.Review, "\n\n")

					reviews = append(reviews, appReviewTemplate{
						Review:     v.Review,
						Player:     player,
						Date:       time.Unix(v.TimestampCreated, 0).Format(helpers.DateYear),
						VotesGood:  v.VotesUp,
						VotesFunny: v.VotesFunny,
						Vote:       v.VotedUp,
					})
				}
			}
		}

		wg.Done()
	}()

	// Wait
	wg.Wait()

	// Template
	t := appTemplate{}
	t.Fill(w, r, app.GetName())
	t.App = app
	t.Packages = packages
	t.Articles = news
	t.Prices = pricesString
	t.PricesCount = pricesCount
	t.Achievements = achievements
	t.Tags = tags
	t.DLC = dlc
	t.Reviews = reviews
	t.ReviewsCount = reviewsCount

	returnTemplate(w, r, "app", t)
}

type appTemplate struct {
	GlobalTemplate
	App          db.App
	Packages     []db.Package
	DLC          []db.App
	Articles     []appArticleTemplate
	Prices       string
	PricesCount  int
	Achievements []appAchievementTemplate
	Schema       steam.SchemaForGame
	Tags         []db.Tag
	Reviews      []appReviewTemplate
	ReviewsCount steam.ReviewsSummaryResponse
}

type appAchievementTemplate struct {
	Icon        string
	Name        string
	Description string
	Completed   float64
}

func (a appAchievementTemplate) GetCompleted() float64 {
	return helpers.DollarsFloat(a.Completed)
}

type appArticleTemplate struct {
	ID       int64
	Title    string
	Contents template.HTML
	Author   string
}

type appReviewTemplate struct {
	Review     string
	Player     db.Player
	Date       string
	VotesGood  int
	VotesFunny int
	Vote       bool
}
