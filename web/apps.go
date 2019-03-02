package web

import (
	"html/template"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/Jleagle/steam-go/steam"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/session"
	"github.com/go-chi/chi"
)

func gamesRouter() http.Handler {

	gamesRedirect := func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, strings.Replace(r.URL.Path, "games", "apps", 1), 302)
	}

	r := chi.NewRouter()
	r.Get("/", gamesRedirect)
	r.Get("/ajax", gamesRedirect)
	r.Get("/{id}", gamesRedirect)
	r.Get("/{id}/ajax/news", gamesRedirect)
	r.Get("/{id}/{slug}", gamesRedirect)
	return r
}

func appsRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", appsHandler)
	r.Get("/ajax", appsAjaxHandler)
	r.Get("/{id}", appHandler)
	r.Get("/{id}/ajax/news", appNewsAjaxHandler)
	r.Get("/{id}/ajax/prices", appPricesAjaxHandler)
	r.Get("/{id}/ajax/players", appPlayersAjaxHandler)
	r.Get("/{id}/ajax/reviews", appReviewsAjaxHandler)
	r.Get("/{id}/{slug}", appHandler)
	return r
}

func appsHandler(w http.ResponseWriter, r *http.Request) {

	// Template
	t := appsTemplate{}
	t.Fill(w, r, "Apps", "") // Description gets set later
	t.Types = db.GetTypesForSelect()
	t.addAssetChosen()
	t.addAssetSlider()

	//
	var wg sync.WaitGroup

	// Get apps count
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Count, err = db.CountApps()
		t.Description = "A live database of " + template.HTML(humanize.Comma(int64(t.Count))) + " Steam games."
		log.Err(err, r)

	}()

	// Get tags
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Tags, err = db.GetTagsForSelect()
		log.Err(err, r)

	}()

	// Get genres
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Genres, err = db.GetGenresForSelect()
		log.Err(err, r)

	}()

	// Get publishers
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Publishers, err = db.GetPublishersForSelect()
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

			publishers, err := db.GetPublishersByID(publishersToLoad, []string{"id", "name"})
			log.Err(err, r)
			if err == nil {
				t.Publishers = append(t.Publishers, publishers...)
			}
		}

	}()

	// Get developers
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Developers, err = db.GetDevelopersForSelect()
		log.Err(err, r)

		// Check if we need to fetch any more to add to the list
		if val, ok := r.URL.Query()["developers"]; ok {

			var developersToLoad []int
			for _, v := range val { // Loop IDs in URL

				// Convert to int
				developerID, err := strconv.Atoi(v)
				if err != nil {
					log.Err(err, r)
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

			developers, err := db.GetDevelopersByID(developersToLoad, []string{"id", "name"})
			log.Err(err, r)
			if err == nil {
				t.Developers = append(t.Developers, developers...)
			}
		}

	}()

	// Get most expensive app
	wg.Add(1)
	go func(r *http.Request) {

		defer wg.Done()

		price, err := db.GetMostExpensiveApp(session.GetCountryCode(r))
		log.Err(err, r)

		// Convert cents to dollars
		t.ExpensiveApp = int(math.Ceil(float64(price) / 100))

		// Fallback
		if t.ExpensiveApp == 0 {
			t.ExpensiveApp = 500
		}

	}(r)

	// Wait
	wg.Wait()

	err := returnTemplate(w, r, "apps", t)
	log.Err(err, r)
}

type appsTemplate struct {
	GlobalTemplate
	Count        int
	ExpensiveApp int
	Types        []db.AppType
	Tags         []db.Tag
	Genres       []db.Genre
	Publishers   []db.Publisher
	Developers   []db.Developer
}

func appsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	setNoCacheHeaders(w)

	query := DataTablesQuery{}
	err := query.FillFromURL(r.URL.Query())
	log.Err(err, r)

	//
	var code = session.GetCountryCode(r)
	var wg sync.WaitGroup

	// Get apps
	var apps []db.App
	var recordsFiltered int

	wg.Add(1)
	go func() {

		defer wg.Done()

		gorm, err := db.GetMySQLClient()
		if err != nil {

			log.Err(err, r)
			return
		}

		gorm = gorm.Model(db.App{})
		gorm = gorm.Select([]string{"id", "name", "icon", "type", "reviews_score", "prices", "change_number_date"})

		// Types
		types := query.GetSearchSlice("types")
		if len(types) > 0 {
			gorm = gorm.Where("type IN (?)", types)
		}

		// Tags
		tags := query.GetSearchSlice("tags")
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
		genres := query.GetSearchSlice("genres")
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
		developers := query.GetSearchSlice("developers")
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
		publishers := query.GetSearchSlice("publishers")
		if len(publishers) > 0 {

			var or []string
			var vals []interface{}
			for _, v := range publishers {
				or = append(or, "JSON_CONTAINS(publishers, ?) = 1")
				vals = append(vals, "["+v+"]")
			}

			gorm = gorm.Where(strings.Join(or, " OR "), vals...)
		}

		// Platforms / Operating System
		platforms := query.GetSearchSlice("platforms")
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
		prices := query.GetSearchSlice("prices")
		if len(prices) == 2 {

			maxPrice, err := db.GetMostExpensiveApp(session.GetCountryCode(r))
			log.Err(err, r)

			// Round up to dollar
			maxPrice = int(math.Ceil(float64(maxPrice)/100) * 100)

			low, err := strconv.Atoi(strings.Replace(prices[0], ".", "", 1))
			log.Err(err, r)

			high, err := strconv.Atoi(strings.Replace(prices[1], ".", "", 1))
			log.Err(err, r)

			var column string
			if code == steam.CountryUS {
				column = "prices_us" // This is an index, just for US
			} else {
				column = "JSON_EXTRACT(prices, \"$." + string(code) + ".final\")"
			}

			if low > 0 {
				gorm = gorm.Where("COALESCE("+column+", 0) >= ?", low)
			}
			if high < maxPrice {
				gorm = gorm.Where("COALESCE("+column+", 0) <= ?", high)
			}

		}

		// Score range
		scores := query.GetSearchSlice("scores")
		if len(scores) == 2 {

			low, err := strconv.Atoi(strings.Replace(scores[0], ".00", "", 1))
			log.Err(err, r)

			high, err := strconv.Atoi(strings.Replace(scores[1], ".00", "", 1))
			log.Err(err, r)

			if low > 0 {
				gorm = gorm.Where("reviews_score >= ?", low)
			}
			if high < 100 {
				gorm = gorm.Where("reviews_score <= ?", high)
			}

		}

		// Search
		search := query.GetSearchString("search")
		if search != "" {
			gorm = gorm.Where("name LIKE ?", "%"+search+"%")
		}

		// Count
		gorm = gorm.Count(&recordsFiltered)
		log.Err(gorm.Error)

		// Order, offset, limit
		gorm = gorm.Limit(100)
		gorm = query.SetOrderOffsetGorm(gorm, code, map[string]string{
			"0": "name",
			"2": "reviews_score",
			"3": "price",
			"4": "change_number_date",
		})

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
		count, err = db.CountApps()
		log.Err(err, r)

	}()

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.Itoa(count)
	response.RecordsFiltered = strconv.Itoa(recordsFiltered)
	response.Draw = query.Draw

	for _, v := range apps {
		response.AddRow(v.OutputForJSON(code))
	}

	response.output(w, r)
}
