package pages

import (
	"html/template"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"

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
	r.Mount("/trending", trendingRouter())
	r.Mount("/{id}", appRouter())
	return r
}

func appsHandler(w http.ResponseWriter, r *http.Request) {

	// Template
	t := appsTemplate{}
	t.fill(w, r, "Apps", "") // Description gets set later
	t.addAssetChosen()
	t.addAssetSlider()

	//
	var wg sync.WaitGroup

	// Get apps count
	wg.Add(1)
	go func() {

		defer wg.Done()

		count, err := mongo.CountDocuments(mongo.CollectionApps, nil, 0)
		t.Description = "A live database of all " + template.HTML(helpers.ShortHandNumber(count)) + " Steam games."
		if err != nil {
			log.Err(err, r)
		}
	}()

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

		var err error
		t.Publishers, err = sql.GetPublishersForSelect()
		log.Err(err, r)

		// Check if we need to fetch any more to add to the list
		if val, ok := r.URL.Query()["publishers"]; ok {

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

		var err error
		t.Developers, err = sql.GetDevelopersForSelect()
		log.Err(err, r)

		// Check if we need to fetch any more to add to the list
		if val, ok := r.URL.Query()["developers"]; ok {

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

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	if err != nil {
		log.Err(err, r)
	}

	//
	var wg sync.WaitGroup
	var code = helpers.GetProductCC(r)
	var filter = bson.D{}
	var countLock sync.Mutex

	// Types
	types := query.getSearchSlice("types")
	if len(types) > 0 {

		a := bson.A{}
		for _, v := range types {
			a = append(a, v)
		}

		filter = append(filter, bson.E{Key: "type", Value: bson.M{"$in": a}})
	}

	// Tags
	tags := query.getSearchSlice("tags")
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
	genres := query.getSearchSlice("genres")
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
	developers := query.getSearchSlice("developers")
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
	publishers := query.getSearchSlice("publishers")
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
	categories := query.getSearchSlice("categories")
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
	platforms := query.getSearchSlice("platforms")
	if len(platforms) > 0 {

		a := bson.A{}
		for _, v := range platforms {
			a = append(a, v)
		}

		filter = append(filter, bson.E{Key: "platforms", Value: bson.M{"$in": a}})
	}

	// Price range
	prices := query.getSearchSlice("price")
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
	scores := query.getSearchSlice("score")
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
	search := query.getSearchString("search")
	if search != "" {
		filter = append(filter, bson.E{Key: "$text", Value: bson.M{"$search": search}})
	}

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

		projection := bson.M{"id": 1, "name": 1, "icon": 1, "reviews_score": 1, "prices": 1, "player_peak_week": 1, "group_followers": 1}
		order := query.getOrderMongo(cols)
		offset := query.getOffset64()

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

	response := DataTablesResponse{}
	response.RecordsTotal = count
	response.RecordsFiltered = recordsFiltered
	response.Draw = query.Draw

	for k, app := range apps {

		response.AddRow([]interface{}{
			app.ID,        // 0
			app.GetName(), // 1
			app.GetIcon(), // 2
			app.GetPath(), // 3
			app.GetType(), // 4
			helpers.RoundFloatTo2DP(app.ReviewsScore), // 5
			app.Prices.Get(code).GetFinal(),           // 6
			app.PlayerPeakWeek,                        // 7
			app.GetStoreLink(),                        // 8
			query.getOffset() + k + 1,                 // 9
			app.GroupFollowers,                        // 10
		})
	}

	response.output(w, r)
}
