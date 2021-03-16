package handlers

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/gamedb/gamedb/cmd/frontend/helpers/datatable"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/session"
	"github.com/go-chi/chi/v5"
	"github.com/olivere/elastic/v7"
)

func GamesRouter() http.Handler {

	r := chi.NewRouter()

	r.Mount("/achievements", appsAchievementsRouter())
	r.Mount("/compare", gamesCompareRouter())
	r.Mount("/coop", coopRouter())
	r.Mount("/dlc", appsDLCRouter())
	r.Mount("/new-releases", newReleasesRouter())
	r.Mount("/release-dates", releaseDatesRouter())
	r.Mount("/sales", salesRouter())
	r.Mount("/trending", trendingRouter())
	r.Mount("/upcoming", upcomingRouter())
	r.Mount("/wallpaper", wallpaperRouter())
	r.Mount("/wishlists", wishlistsRouter())
	r.Mount("/{id:[0-9]+}", appRouter())

	r.Get("/", appsHandler)
	r.Get("/games.json", appsAjaxHandler)
	r.Get("/random", appsRandomHandler)

	return r
}

func appsHandler(w http.ResponseWriter, r *http.Request) {

	// Template
	t := appsTemplate{}
	t.fill(w, r, "apps", "Games", "A live database of all Steam games")
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
			log.ErrS(err)
		}
	}()

	// Get tags
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Tags, err = mongo.GetStatsForSelect(mongo.StatsTypeTags)
		if err != nil {
			log.ErrS(err)
		}
	}()

	// Get genres
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Genres, err = mongo.GetStatsForSelect(mongo.StatsTypeGenres)
		if err != nil {
			log.ErrS(err)
		}
	}()

	// Get categories
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Categories, err = mongo.GetStatsForSelect(mongo.StatsTypeCategories)
		if err != nil {
			log.ErrS(err)
		}
	}()

	// Get publishers
	wg.Add(1)
	go func() {

		defer wg.Done()

		if val, ok := r.URL.Query()["publishers"]; ok {

			var err error
			t.Publishers, err = mongo.GetStatsForSelect(mongo.StatsTypePublishers)
			if err != nil {
				log.ErrS(err)
			}

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

			publishers, err := mongo.GetStatsByID(mongo.StatsTypePublishers, publishersToLoad)
			if err != nil {
				log.ErrS(err)
			} else {
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
			t.Developers, err = mongo.GetStatsForSelect(mongo.StatsTypeDevelopers)
			if err != nil {
				log.ErrS(err)
			}

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

			developers, err := mongo.GetStatsByID(mongo.StatsTypeDevelopers, developersToLoad)
			if err != nil {
				log.ErrS(err)
			} else {
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

	returnTemplate(w, r, t)
}

type appsTemplate struct {
	globalTemplate
	Types      []mongo.AppTypeCount
	Tags       []mongo.Stat
	Genres     []mongo.Stat
	Categories []mongo.Stat
	Publishers []mongo.Stat
	Developers []mongo.Stat
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

		low, err := strconv.Atoi(strings.Replace(prices[0], ".", "", 1))
		if err == nil && low > 0 {
			filters = append(filters, elastic.NewRangeQuery("prices."+string(code)+".final").From(low))
		}

		high, err := strconv.Atoi(strings.Replace(prices[1], ".", "", 1))
		if err == nil && high < 100_00 {
			filters = append(filters, elastic.NewRangeQuery("prices."+string(code)+".final").To(high))
		}
	}

	scores := query.GetSearchSlice("score")
	if len(scores) == 2 {

		low, err := strconv.Atoi(scores[0])
		if err == nil && low > 0 {
			filters = append(filters, elastic.NewRangeQuery("score").From(low))
		}

		high, err := strconv.Atoi(scores[1])
		if err == nil && high < 100 {
			filters = append(filters, elastic.NewRangeQuery("score").To(high))
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
		apps, recordsFiltered, err = elasticsearch.SearchAppsAdvanced(query.GetOffset(), 100, search, order, elastic.NewBoolQuery().Filter(filters...))
		if err != nil {
			log.ErrS(err)
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
			log.ErrS(err)
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
