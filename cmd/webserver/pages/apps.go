package pages

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/datatable"
	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/go-chi/chi"
	"github.com/olivere/elastic/v7"
)

func GamesRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", appsHandler)
	r.Get("/games.json", appsAjaxHandler)
	r.Get("/random", appsRandomHandler)
	r.Mount("/coop", coopRouter())
	r.Mount("/compare", gamesCompareRouter())
	r.Mount("/upcoming", upcomingRouter())
	r.Mount("/achievements", appsAchievementsRouter())
	r.Mount("/new-releases", newReleasesRouter())
	r.Mount("/sales", salesRouter())
	r.Mount("/trending", trendingRouter())
	r.Mount("/wishlists", wishlistsRouter())
	r.Mount("/wallpaper", WallpaperRouter())
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

		var code = session.GetProductCC(r)

		var err error
		t.Types, err = mongo.GetAppsGroupedByType(code)
		if err != nil {
			log.Err(err, r)
		}
	}()

	// Get tags
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Tags, err = mysql.GetTagsForSelect()
		log.Err(err, r)
	}()

	// Get genres
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Genres, err = mysql.GetGenresForSelect()
		log.Err(err, r)
	}()

	// Get categories
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Categories, err = mysql.GetCategoriesForSelect()
		log.Err(err, r)
	}()

	// Get publishers
	wg.Add(1)
	go func() {

		defer wg.Done()

		if val, ok := r.URL.Query()["publishers"]; ok {

			var err error
			t.Publishers, err = mysql.GetPublishersForSelect()
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

			publishers, err := mysql.GetPublishersByID(publishersToLoad, []string{"id", "name"})
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
			t.Developers, err = mysql.GetDevelopersForSelect()
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

			developers, err := mysql.GetDevelopersByID(developersToLoad, []string{"id", "name"})
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
	globalTemplate
	Types      []mongo.AppTypeCount
	Tags       []mysql.Tag
	Genres     []mysql.Genre
	Categories []mysql.Category
	Publishers []mysql.Publisher
	Developers []mysql.Developer
}

func appsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := datatable.NewDataTableQuery(r, false)

	//
	var wg sync.WaitGroup
	var code = session.GetProductCC(r)
	var filters []elastic.Query

	types := query.GetSearchSliceInterface("types")
	if len(types) > 0 {
		filters = append(filters, elastic.NewTermsQuery("type", types...))
	}

	tags := query.GetSearchSliceInterface("tags")
	if len(tags) > 0 {
		filters = append(filters, elastic.NewTermsQuery("tags", tags...))
	}

	genres := query.GetSearchSliceInterface("genres")
	if len(genres) > 0 {
		filters = append(filters, elastic.NewTermsQuery("genres", genres...))
	}

	developers := query.GetSearchSliceInterface("developers")
	if len(developers) > 0 {
		filters = append(filters, elastic.NewTermsQuery("developers", developers...))
	}

	publishers := query.GetSearchSliceInterface("publishers")
	if len(publishers) > 0 {
		filters = append(filters, elastic.NewTermsQuery("publishers", publishers...))
	}

	categories := query.GetSearchSliceInterface("categories")
	if len(categories) > 0 {
		filters = append(filters, elastic.NewTermsQuery("categories", categories...))
	}

	platforms := query.GetSearchSliceInterface("platforms")
	if len(platforms) > 0 {
		filters = append(filters, elastic.NewTermsQuery("platforms", platforms...))
	}

	prices := query.GetSearchSlice("price")
	if len(prices) == 2 {

		lowCheck, highCheck := false, false

		q := elastic.NewRangeQuery("prices." + string(code) + ".final")

		low, err := strconv.Atoi(strings.Replace(prices[0], ".", "", 1))
		log.Err(err, r)
		if err == nil && low > 0 {
			lowCheck = true
			q.From(low)
		}

		high, err := strconv.Atoi(strings.Replace(prices[1], ".", "", 1))
		log.Err(err, r)
		if err == nil && high < 100*100 {
			highCheck = true
			q.To(high)
		}

		if lowCheck || highCheck {
			filters = append(filters, q)
		}
	}

	scores := query.GetSearchSlice("score")
	if len(scores) == 2 {

		lowCheck, highCheck := false, false

		q := elastic.NewRangeQuery("score")

		low, err := strconv.Atoi(strings.TrimSuffix(scores[0], ".00"))
		log.Err(err, r)
		if err == nil && low > 0 {
			lowCheck = true
			q.From(low)
		}

		high, err := strconv.Atoi(strings.TrimSuffix(scores[1], ".00"))
		log.Err(err, r)
		if err == nil && high < 100 {
			highCheck = true
			q.To(high)
		}

		if lowCheck || highCheck {
			filters = append(filters, q)
		}
	}

	// Get apps
	var apps []elasticsearch.App
	var recordsFiltered int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		cols := map[string]string{
			"2": "players",
			"3": "followers",
			"4": "score",
			"5": "prices." + string(code) + ".final",
		}

		order := query.GetOrderElastic(cols)
		search := query.GetSearchString("search")

		var err error
		apps, recordsFiltered, err = elasticsearch.SearchAppsAdvanced(query.GetOffset(), search, order, filters)
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
		count, err = mongo.CountDocuments(mongo.CollectionApps, nil, 0)
		if err != nil {
			log.Err(err, r)
		}
	}()

	// Wait
	wg.Wait()

	var response = datatable.NewDataTablesResponse(r, query, count, recordsFiltered, nil)
	for k, app := range apps {

		var formattedReviewScore = helpers.RoundFloatTo2DP(app.ReviewScore)

		response.AddRow([]interface{}{
			app.ID,                          // 0
			app.GetName(),                   // 1
			app.GetIcon(),                   // 2
			app.GetPath(),                   // 3
			nil,                             // 4
			formattedReviewScore,            // 5
			app.Prices.Get(code).GetFinal(), // 6
			app.PlayersCount,                // 7
			app.GetStoreLink(),              // 8
			query.GetOffset() + k + 1,       // 9
			app.FollowersCount,              // 10
			app.Score,                       // 11
			app.GetMarkedName(),             // 12
		})
	}

	returnJSON(w, r, response)
}
