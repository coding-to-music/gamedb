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
	"github.com/go-chi/chi"
	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/helpers"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
	"github.com/steam-authority/steam-authority/steami"
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

	if !mysql.IsValidAppID(idx) {
		returnErrorTemplate(w, r, 400, "Invalid App ID: "+id)
		return
	}

	// Get app
	app, err := mysql.GetApp(idx)
	if err != nil {

		if err.Error() == "no id" {
			returnErrorTemplate(w, r, 404, "We can't find this app in our database, there may not be one with this ID.")
			return
		}

		logger.Error(err)
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
		logger.Error(v)
	}

	//
	var wg sync.WaitGroup

	var achievements []appAchievementTemplate
	wg.Add(1)
	go func() {

		// Get achievements
		achievementsResp,_, err := steami.Steam().GetGlobalAchievementPercentagesForApp(app.ID)
		if err != nil {
			logger.Error(err)
			return
		}

		achievementsMap := make(map[string]float64)
		for _, v := range achievementsResp.GlobalAchievementPercentage {
			achievementsMap[v.Name] = v.Percent
		}

		// Get schema
		schema, _, err := steami.Steam().GetSchemaForGame(app.ID)
		if err != nil {
			logger.Error(err)
			return
		}

		// Make template struct
		for _, v := range schema.AvailableGameStats.Achievements {
			achievements = append(achievements, appAchievementTemplate{
				v.Icon,
				v.DisplayName,
				v.Description,
				achievementsMap[v.Name],
			})
		}

		wg.Done()
	}()

	var tags []mysql.Tag
	wg.Add(1)
	go func() {

		// Get tags
		tagIDs, err := app.GetTags()
		if err != nil {
			logger.Error(err)
			return
		}

		tags, err = mysql.GetTagsByID(tagIDs)
		if err != nil {
			logger.Error(err)
			return
		}

		wg.Done()
	}()

	var pricesString string
	var pricesCount int
	wg.Add(1)
	go func() {

		// Get prices
		pricesResp, err := datastore.GetAppPrices(app.ID, 0)
		if err != nil {
			logger.Error(err)
			return
		}

		pricesCount = len(pricesResp)

		var prices [][]float64

		for _, v := range pricesResp {

			prices = append(prices, []float64{float64(v.CreatedAt.Unix()), float64(v.PriceFinal) / 100})
		}

		// Add current price
		prices = append(prices, []float64{float64(time.Now().Unix()), float64(app.PriceFinal) / 100})

		// Make into a JSON string
		pricesBytes, err := json.Marshal(prices)
		if err != nil {
			logger.Error(err)
			return
		}

		pricesString = string(pricesBytes)

		wg.Done()
	}()

	var news []appArticleTemplate
	wg.Add(1)
	go func() {

		// Get news
		newsResp, err := datastore.GetArticles(idx, 1000)
		if err != nil {
			logger.Error(err)
			return
		}

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
				Contents: template.HTML(v.Contents),
				Author:   v.Author,
			})
		}

		wg.Done()
	}()

	var packages []mysql.Package
	wg.Add(1)
	go func() {

		// Get packages
		packages, err = mysql.GetPackagesAppIsIn(app.ID)
		if err != nil {
			logger.Error(err)
			return
		}

		wg.Done()
	}()

	var dlc []mysql.App
	wg.Add(1)
	go func() {

		// Get DLC
		dlc, err = mysql.GetDLC(app, []string{"id", "name"})
		if err != nil {
			logger.Error(err)
			return
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
			logger.Error(err)
			return
		}

		reviewsCount = reviewsResponse.QuerySummary

		// Make slice of playerIDs
		var playerIDs []int64
		for _, v := range reviewsResponse.Reviews {
			playerIDs = append(playerIDs, v.Author.SteamID)
		}

		players, err := datastore.GetPlayersByIDs(playerIDs)
		if err != nil {
			logger.Error(err)
			return
		}

		// Make map of players
		var playersMap = map[int64]datastore.Player{}
		for _, v := range players {
			playersMap[v.PlayerID] = v
		}

		// Make template slice
		for _, v := range reviewsResponse.Reviews {

			var player datastore.Player
			if val, ok := playersMap[v.Author.SteamID]; ok {
				player = val
			} else {
				player = datastore.Player{}
				player.PlayerID = v.Author.SteamID
				player.PersonaName = "Unknown"
			}

			// Remove extra new lines
			regex := regexp.MustCompile("[\n]{3,}") // After comma
			v.Review = regex.ReplaceAllString(v.Review, "\n\n")

			reviews = append(reviews, appReviewTemplate{
				Review: v.Review,
				Player: player,
				Date:   time.Unix(v.TimestampCreated, 0).Format(helpers.DateYear),
				Votes:  v.VotesUp,
			})
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
	App          mysql.App
	Packages     []mysql.Package
	DLC          []mysql.App
	Articles     []appArticleTemplate
	Prices       string
	PricesCount  int
	Achievements []appAchievementTemplate
	Schema       steam.SchemaForGame
	Tags         []mysql.Tag
	Reviews      []appReviewTemplate
	ReviewsCount steam.ReviewsSummaryResponse
}

type appAchievementTemplate struct {
	Icon        string
	Name        string
	Description string
	Completed   float64
}

type appArticleTemplate struct {
	ID       int64
	Title    string
	Contents template.HTML
	Author   string
}

type appReviewTemplate struct {
	Review string
	Player datastore.Player
	Date   string
	Votes  int
}

func (a appAchievementTemplate) GetCompleted() float64 {
	return helpers.DollarsFloat(a.Completed)
}

type hcSeries struct {
	Type string
	Name string
	Data [][]int64
	Step bool
}
