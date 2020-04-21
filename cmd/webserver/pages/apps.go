package pages

import (
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/gamedb/gamedb/cmd/webserver/helpers/datatable"
	"github.com/gamedb/gamedb/cmd/webserver/helpers/session"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
)

func AppsRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", appsHandler)
	r.Get("/apps.json", appsAjaxHandler)
	r.Get("/random", appsRandomHandler)
	r.Mount("/trending", trendingRouter())
	r.Mount("/{id}", appRouter())
	return r
}

func appsHandler(w http.ResponseWriter, r *http.Request) {

	// Template
	t := appsTemplate{}
	t.fill(w, r, "Games", "A live database of all Steam games")
	t.addAssetChosen()
	t.addAssetSlider()

	//
	var wg sync.WaitGroup

	// Get apps types
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Types, err = mongo.GetAppTypes()
		if err != nil {
			log.Err(err, r)
		}
	}()

	// Get tags
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Tags, err = sql.GetTagsForSelect()
		log.Err(err, r)
	}()

	// Get genres
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Genres, err = sql.GetGenresForSelect()
		log.Err(err, r)
	}()

	// Get categories
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Categories, err = sql.GetCategoriesForSelect()
		log.Err(err, r)
	}()

	// Get publishers
	wg.Add(1)
	go func() {

		defer wg.Done()

		if val, ok := r.URL.Query()["publishers"]; ok {

			var err error
			t.Publishers, err = sql.GetPublishersForSelect()
			log.Err(err, r)

			var publishersToLoad []int
			for _, v := range val { // Loop IDs in URL

				// Convert to int
				publisherID, err := strconv.Atoi(v)
				if err != nil {
					continue // No need to log invalid values
				}

				// Check if we already have this publisher
				var alreadyHavePublisher = false
				for _, vv := range t.Publishers {
					if publisherID == vv.ID {
						alreadyHavePublisher = true
						break
					}
				}

				// Add to slice to load
				if !alreadyHavePublisher {
					publishersToLoad = append(publishersToLoad, publisherID)
				}
			}

			publishers, err := sql.GetPublishersByID(publishersToLoad, []string{"id", "name"})
			log.Err(err, r)
			if err == nil {
				t.Publishers = append(t.Publishers, publishers...)
			}
		}

		sort.Slice(t.Publishers, func(i, j int) bool {
			return strings.ToLower(t.Publishers[i].Name) < strings.ToLower(t.Publishers[j].Name)
		})
	}()

	// Get developers
	wg.Add(1)
	go func() {

		defer wg.Done()

		if val, ok := r.URL.Query()["developers"]; ok {

			var err error
			t.Developers, err = sql.GetDevelopersForSelect()
			log.Err(err, r)

			var developersToLoad []int
			for _, v := range val { // Loop IDs in URL

				// Convert to int
				developerID, err := strconv.Atoi(v)
				if err != nil {
					continue // No need to log invalid values
				}

				// Check if we already have this developer
				var alreadyHaveDeveloper = false
				for _, vv := range t.Developers {
					if developerID == vv.ID {
						alreadyHaveDeveloper = true
						break
					}
				}

				// Add to slice to load
				if !alreadyHaveDeveloper {
					developersToLoad = append(developersToLoad, developerID)
				}
			}

			developers, err := sql.GetDevelopersByID(developersToLoad, []string{"id", "name"})
			log.Err(err, r)
			if err == nil {
				t.Developers = append(t.Developers, developers...)
			}
		}

		sort.Slice(t.Developers, func(i, j int) bool {
			return strings.ToLower(t.Developers[i].Name) < strings.ToLower(t.Developers[j].Name)
		})
	}()

	// Wait
	wg.Wait()

	// t.Columns = allColumns

	returnTemplate(w, r, "apps", t)
}

type appsTemplate struct {
	GlobalTemplate
	Types      []mongo.AppTypeCount
	Tags       []sql.Tag
	Genres     []sql.Genre
	Categories []sql.Category
	Publishers []sql.Publisher
	Developers []sql.Developer
}

func appsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := datatable.NewDataTableQuery(r, false)

	//
	var wg sync.WaitGroup
	var code = session.GetProductCC(r)
	var filter = bson.D{}
	var countLock sync.Mutex

	// Types
	types := query.GetSearchSlice("types")
	if len(types) > 0 {

		a := bson.A{}
		for _, v := range types {
			a = append(a, v)
		}

		filter = append(filter, bson.E{Key: "type", Value: bson.M{"$in": a}})
	}

	// Tags
	tags := query.GetSearchSlice("tags")
	if len(tags) > 0 {

		a := bson.A{}
		for _, v := range tags {
			i, err := strconv.Atoi(v)
			if err == nil {
				a = append(a, i)
			}
		}

		filter = append(filter, bson.E{Key: "tags", Value: bson.M{"$in": a}})
	}

	// Genres
	genres := query.GetSearchSlice("genres")
	if len(genres) > 0 {

		a := bson.A{}
		for _, v := range genres {
			i, err := strconv.Atoi(v)
			if err == nil {
				a = append(a, i)
			}
		}

		filter = append(filter, bson.E{Key: "genres", Value: bson.M{"$in": a}})
	}

	// Developers
	developers := query.GetSearchSlice("developers")
	if len(developers) > 0 {

		a := bson.A{}
		for _, v := range developers {
			i, err := strconv.Atoi(v)
			if err == nil {
				a = append(a, i)
			}
		}

		filter = append(filter, bson.E{Key: "developers", Value: bson.M{"$in": a}})
	}

	// Publishers
	publishers := query.GetSearchSlice("publishers")
	if len(publishers) > 0 {

		a := bson.A{}
		for _, v := range publishers {
			i, err := strconv.Atoi(v)
			if err == nil {
				a = append(a, i)
			}
		}

		filter = append(filter, bson.E{Key: "publishers", Value: bson.M{"$in": a}})
	}

	// Categories
	categories := query.GetSearchSlice("categories")
	if len(categories) > 0 {

		a := bson.A{}
		for _, v := range categories {
			i, err := strconv.Atoi(v)
			if err == nil {
				a = append(a, i)
			}
		}

		filter = append(filter, bson.E{Key: "categories", Value: bson.M{"$in": a}})
	}

	// Platforms
	platforms := query.GetSearchSlice("platforms")
	if len(platforms) > 0 {

		a := bson.A{}
		for _, v := range platforms {
			a = append(a, v)
		}

		filter = append(filter, bson.E{Key: "platforms", Value: bson.M{"$in": a}})
	}

	// Price range
	prices := query.GetSearchSlice("price")
	if len(prices) == 2 {

		low, err := strconv.Atoi(strings.Replace(prices[0], ".", "", 1))
		log.Err(err, r)

		high, err := strconv.Atoi(strings.Replace(prices[1], ".", "", 1))
		log.Err(err, r)

		var column = "prices." + string(code) + ".final"

		if low > 0 {
			filter = append(filter, bson.E{Key: column, Value: bson.M{"$gte": low}})
		}
		if high < 100*100 {
			filter = append(filter, bson.E{Key: column, Value: bson.M{"$lte": high}})
		}
	}

	// Score range
	scores := query.GetSearchSlice("score")
	if len(scores) == 2 {

		low, err := strconv.Atoi(strings.TrimSuffix(scores[0], ".00"))
		log.Err(err, r)

		high, err := strconv.Atoi(strings.TrimSuffix(scores[1], ".00"))
		log.Err(err, r)

		if low > 0 {
			filter = append(filter, bson.E{Key: "reviews_score", Value: bson.M{"$gte": low}})
		}
		if high < 100 {
			filter = append(filter, bson.E{Key: "reviews_score", Value: bson.M{"$lte": high}})
		}
	}

	// Search
	search := query.GetSearchString("search")
	if search != "" {

		if !strings.Contains(search, `"`) {
			search = regexp.MustCompile(`([^\s]+)`).ReplaceAllString(search, `"$1"`) // Add quotes to all words
		}

		filter = append(filter, bson.E{Key: "$text", Value: bson.M{"$search": search}})
	}

	// Temp
	filter = append(filter, bson.E{Key: "_id", Value: bson.M{"$gt": 0}})

	// Get apps
	var apps []mongo.App
	wg.Add(1)
	go func() {

		defer wg.Done()

		cols := map[string]string{
			"2": "player_peak_week",
			"3": "group_followers",
			"4": "reviews_score",
			"5": "prices." + string(code) + ".final",
		}

		projection := bson.M{"_id": 1, "name": 1, "icon": 1, "reviews_score": 1, "prices": 1, "player_peak_week": 1, "group_followers": 1}
		order := query.GetOrderMongo(cols)
		offset := query.GetOffset64()

		var err error
		apps, err = mongo.GetApps(offset, 100, order, filter, projection, nil)
		if err != nil {
			log.Err(err, r)
		}
	}()

	// Get filtered count
	var recordsFiltered int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		countLock.Lock()
		recordsFiltered, err = mongo.CountDocuments(mongo.CollectionApps, filter, 10)
		countLock.Unlock()
		if err != nil {
			log.Err(err, r)
		}
	}()

	// Get count
	var count int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		countLock.Lock()
		count, err = mongo.CountDocuments(mongo.CollectionApps, nil, 0)
		countLock.Unlock()
		if err != nil {
			log.Err(err, r)
		}
	}()

	// Wait
	wg.Wait()

	var response = datatable.NewDataTablesResponse(r, query, count, recordsFiltered)
	for k, app := range apps {

		var formattedReviewScore = helpers.RoundFloatTo2DP(app.ReviewsScore)

		response.AddRow([]interface{}{
			app.ID,                          // 0
			app.GetName(),                   // 1
			app.GetIcon(),                   // 2
			app.GetPath(),                   // 3
			app.GetType(),                   // 4
			formattedReviewScore,            // 5
			app.Prices.Get(code).GetFinal(), // 6
			app.PlayerPeakWeek,              // 7
			app.GetStoreLink(),              // 8
			query.GetOffset() + k + 1,       // 9
			app.GroupFollowers,              // 10
		})
	}

	returnJSON(w, r, response)
}

func appsRandomHandler(w http.ResponseWriter, r *http.Request) {

	var t = appsRandomTemplate{}

	var player mongo.Player

	var filter = bson.D{
		{"name", bson.M{"$ne": ""}},
		{"type", "game"},
	}

	if session.IsLoggedIn(r) {

		ids := bson.A{}

		user, err := getUserFromSession(r)
		if err != nil {
			log.Err(err)
			returnErrorTemplate(w, r, errorTemplate{Code: 500})
			return
		}

		var steamID = user.GetSteamID()
		if steamID > 0 {

			player, err = mongo.GetPlayer(steamID)
			if err != nil {
				log.Err(err)
				returnErrorTemplate(w, r, errorTemplate{Code: 500})
				return
			}

			playerApps, err := mongo.GetPlayerApps(steamID, 0, 0, nil)
			if err != nil {
				log.Err(err)
				returnErrorTemplate(w, r, errorTemplate{Code: 500})
				return
			}

			for _, v := range playerApps {
				ids = append(ids, v.AppID)
			}

			filter = append(filter, bson.E{Key: "_id", Value: bson.M{"$in": ids}})
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
	}

	apps, err := mongo.GetRandomApps(1, filter, projection)
	if err != nil {
		log.Err(err)
		returnErrorTemplate(w, r, errorTemplate{Code: 500})
		return
	}
	if len(apps) == 0 {
		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "Couldn't find a game"})
		return
	}

	t.setBackground(apps[0], false, false)
	t.fill(w, r, "Random Steam Game", "Find a random Steam game")
	t.addAssetChosen()
	t.addAssetSlider()

	t.Apps = apps
	t.Player = player

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Tags, err = sql.GetTagsForSelect()
		log.Err(err, r)
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		if len(apps) > 0 {

			var err error
			t.AppTags, err = GetAppTags(apps[0])
			if err != nil {
				log.Err(err, r)
			}
		}
	}()

	wg.Wait()

	returnTemplate(w, r, "apps_random", t)
}

type appsRandomTemplate struct {
	GlobalTemplate
	Apps    []mongo.App
	Player  mongo.Player
	Tags    []sql.Tag
	AppTags []sql.Tag
}
