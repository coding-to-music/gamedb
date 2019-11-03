package pages

import (
	"html/template"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/go-chi/chi"
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
	t.Types = sql.GetTypesForSelect()
	t.addAssetChosen()
	t.addAssetSlider()

	//
	var wg sync.WaitGroup

	// Get apps count
	wg.Add(1)
	go func() {

		defer wg.Done()

		count, err := sql.CountApps()
		t.Description = "A live database of all " + template.HTML(helpers.ShortHandNumber(int64(count))) + " Steam games."
		log.Err(err, r)

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
					log.Err(err, r)
					continue
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
					log.Info(err, r)
					continue
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
	Types      []sql.AppType
	Tags       []sql.Tag
	Genres     []sql.Genre
	Categories []sql.Category
	Publishers []sql.Publisher
	Developers []sql.Developer
}

func appsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	//
	var code = helpers.GetProductCC(r)
	var wg sync.WaitGroup

	// Get apps
	var apps []sql.App
	var recordsFiltered int64

	wg.Add(1)
	go func() {

		defer wg.Done()

		gorm, err := sql.GetMySQLClient()
		if err != nil {

			log.Err(err, r)
			return
		}

		gorm = gorm.Model(sql.App{})
		gorm = gorm.Select([]string{"id", "name", "icon", "reviews_score", "prices", "player_peak_week"})

		// Types
		types := query.getSearchSlice("types")
		if len(types) > 0 {
			gorm = gorm.Where("type IN (?)", types)
		}

		// Tags
		tags := query.getSearchSlice("tags")
		if len(tags) > 0 {

			var or []string
			var vals []interface{}
			for _, v := range tags {
				or = append(or, "JSON_CONTAINS(tags, ?) = 1")
				vals = append(vals, "["+v+"]")
			}

			gorm = gorm.Where(strings.Join(or, " OR "), vals...)
		}

		// Genres
		genres := query.getSearchSlice("genres")
		if len(genres) > 0 {

			var or []string
			var vals []interface{}
			for _, v := range genres {
				or = append(or, "JSON_CONTAINS(genres, ?) = 1")
				vals = append(vals, "["+v+"]")
			}

			gorm = gorm.Where(strings.Join(or, " OR "), vals...)
		}

		// Developers
		developers := query.getSearchSlice("developers")
		if len(developers) > 0 {

			var or []string
			var vals []interface{}
			for _, v := range developers {
				or = append(or, "JSON_CONTAINS(developers, ?) = 1")
				vals = append(vals, "["+v+"]")
			}

			gorm = gorm.Where(strings.Join(or, " OR "), vals...)
		}

		// Publishers
		publishers := query.getSearchSlice("publishers")
		if len(publishers) > 0 {

			var or []string
			var vals []interface{}
			for _, v := range publishers {
				or = append(or, "JSON_CONTAINS(publishers, ?) = 1")
				vals = append(vals, "["+v+"]")
			}

			gorm = gorm.Where(strings.Join(or, " OR "), vals...)
		}

		// Categories
		categories := query.getSearchSlice("categories")
		if len(categories) > 0 {

			var or []string
			var vals []interface{}
			for _, v := range categories {
				or = append(or, "JSON_CONTAINS(categories, ?) = 1")
				vals = append(vals, "["+v+"]")
			}

			gorm = gorm.Where(strings.Join(or, " OR "), vals...)
		}

		// Platforms / Operating System
		platforms := query.getSearchSlice("platforms")
		if len(platforms) > 0 {

			var or []string
			var vals []interface{}
			for _, v := range platforms {
				or = append(or, "JSON_CONTAINS(platforms, ?) = 1")
				vals = append(vals, "[\""+v+"\"]")
			}
			gorm = gorm.Where(strings.Join(or, " OR "), vals...)
		}

		// Price range
		prices := query.getSearchSlice("price")
		if len(prices) == 2 {

			low, err := strconv.Atoi(strings.Replace(prices[0], ".", "", 1))
			log.Err(err, r)

			high, err := strconv.Atoi(strings.Replace(prices[1], ".", "", 1))
			log.Err(err, r)

			var column string

			switch code {
			case steam.ProductCCUS, steam.ProductCCUK, steam.ProductCCEU: // Indexed columns
				column = "prices_" + string(code)
			default:
				column = "JSON_EXTRACT(prices, \"$." + string(code) + ".final\")"
			}

			if low > 0 {
				gorm = gorm.Where("COALESCE("+column+", 0) >= ?", low)
			}
			if high < 100*100 {
				gorm = gorm.Where("COALESCE("+column+", 0) <= ?", high)
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
				gorm = gorm.Where("reviews_score >= ?", low)
			}
			if high < 100 {
				gorm = gorm.Where("reviews_score <= ?", high)
			}
		}

		// Search
		search := query.getSearchString("search")
		if search != "" {
			gorm = gorm.Where("name LIKE ?", "%"+search+"%")
		}

		// Count
		gorm = gorm.Count(&recordsFiltered)
		log.Err(gorm.Error)

		// Order, offset, limit
		cols := map[string]string{
			"2": "player_peak_week",
			"3": "reviews_score",
			"4": "JSON_EXTRACT(prices, \"$." + string(code) + ".final\")",
		}
		gorm = query.setOrderOffsetGorm(gorm, cols, "2")
		gorm = gorm.Limit(100)

		// Get rows
		gorm = gorm.Find(&apps)
		log.Err(gorm.Error)
	}()

	// Get total
	var count int
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		count, err = sql.CountApps()
		log.Err(err, r)

	}()

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = int64(count)
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
			app.GetPrice(code).GetFinal(),             // 6
			app.PlayerPeakWeek,                        // 7
			app.GetStoreLink(),                        // 8
			query.getOffset() + k + 1,                 // 9
		})
	}

	response.output(w, r)
}
