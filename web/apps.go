package web

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/Masterminds/squirrel"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/session"
	"github.com/go-chi/chi"
)

func gamesRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", AppsHandler)
	r.Get("/ajax", AppsAjaxHandler)
	r.Get("/{id}", AppHandler)
	r.Get("/{id}/ajax/news", AppNewsAjaxHandler)
	r.Get("/{id}/{slug}", AppHandler)
	return r
}

func AppsHandler(w http.ResponseWriter, r *http.Request) {

	// Template
	t := appsTemplate{}
	t.Fill(w, r, "Games")
	t.Types = db.GetTypesForSelect()

	//
	var wg sync.WaitGroup

	// Get apps count
	wg.Add(1)
	go func() {

		var err error
		t.Count, err = db.CountApps()
		t.Description = "A live database of " + humanize.Comma(int64(t.Count)) + " Steam games."
		log.Log(err)

		wg.Done()
	}()

	// Get tags
	wg.Add(1)
	go func() {

		var err error
		t.Tags, err = db.GetTagsForSelect()
		log.Log(err)

		wg.Done()
	}()

	// Get genres
	wg.Add(1)
	go func() {

		var err error
		t.Genres, err = db.GetGenresForSelect()
		log.Log(err)

		wg.Done()
	}()

	// Get publishers
	wg.Add(1)
	go func() {

		var err error
		t.Publishers, err = db.GetPublishersForSelect()
		log.Log(err)

		if val, ok := r.URL.Query()["publishers"]; ok {

			var publishersToLoad []int
			for _, v := range val { // Loop IDs in URL

				// Convert to int
				publisherID, err := strconv.Atoi(v)
				if err != nil {
					log.Log(err)
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
			log.Log(err)
			if err == nil {
				for _, vvv := range publishers {
					t.Publishers = append(t.Publishers, vvv)
				}
			}
		}

		wg.Done()
	}()

	// Get developers
	wg.Add(1)
	go func() {

		var err error
		t.Developers, err = db.GetDevelopersForSelect()
		log.Log(err)

		if val, ok := r.URL.Query()["developers"]; ok {

			var developersToLoad []int
			for _, v := range val { // Loop IDs in URL

				// Convert to int
				developerID, err := strconv.Atoi(v)
				if err != nil {
					log.Log(err)
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
			log.Log(err)
			if err == nil {
				for _, vvv := range developers {
					t.Developers = append(t.Developers, vvv)
				}
			}
		}

		wg.Done()
	}()

	// Get most expensive app
	wg.Add(1)
	go func(r *http.Request) {

		price, err := db.GetMostExpensiveApp(session.GetCountryCode(r))
		log.Log(err)

		// Convert cents to dollars
		t.ExpensiveApp = int(math.Ceil(float64(price) / 100))

		wg.Done()
	}(r)

	// Wait
	wg.Wait()

	err := returnTemplate(w, r, "apps", t)
	log.Log(err)
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

func AppsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	setNoCacheHeaders(w)

	query := DataTablesQuery{}
	err := query.FillFromURL(r.URL.Query())
	log.Log(err)

	//
	var code = session.GetCountryCode(r)
	var wg sync.WaitGroup

	// Get apps
	var apps []db.App
	var recordsFiltered int

	wg.Add(1)
	go func() {

		gorm, err := db.GetMySQLClient()
		if err != nil {

			log.Log(err)

		} else {

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

				var or squirrel.Or
				for _, v := range tags {
					or = append(or, squirrel.Eq{"JSON_CONTAINS(tags, '[" + v + "]')": 1})
				}
				sql, data, err := or.ToSql()
				log.Log(err)

				gorm = gorm.Where(sql, data)
			}

			// Genres
			genres := query.GetSearchSlice("genres")
			if len(genres) > 0 {

				var or squirrel.Or
				for _, v := range genres {
					or = append(or, squirrel.Eq{"JSON_CONTAINS(genres, JSON_OBJECT('id', " + v + "))": 1})
				}
				sql, data, err := or.ToSql()
				log.Log(err)

				gorm = gorm.Where(sql, data...)
			}

			// Developers
			developers := query.GetSearchSlice("developers")
			if len(developers) > 0 {

				var or squirrel.Or
				for _, v := range developers {
					or = append(or, squirrel.Eq{"JSON_CONTAINS(developers, '[\"" + v + "\"]')": 1})
				}
				sql, data, err := or.ToSql()
				log.Log(err)

				gorm = gorm.Where(sql, data...)
			}

			// Publishers
			publishers := query.GetSearchSlice("publishers")
			if len(publishers) > 0 {

				var or squirrel.Or
				for _, v := range publishers {
					or = append(or, squirrel.Eq{"JSON_CONTAINS(publishers, '[\"" + v + "\"]')": 1})
				}
				sql, data, err := or.ToSql()
				log.Log(err)

				gorm = gorm.Where(sql, data...)
			}

			// Platforms
			platforms := query.GetSearchSlice("platforms")
			if len(platforms) > 0 {

				var or squirrel.Or
				for _, v := range platforms {
					or = append(or, squirrel.Eq{"JSON_CONTAINS(platforms, '[\"" + v + "\"]')": 1})
				}
				sql, data, err := or.ToSql()
				log.Log(err)

				gorm = gorm.Where(sql, data...)
			}

			// Price range
			prices := query.GetSearchSlice("prices")
			if len(prices) == 2 {

				high, err := strconv.Atoi(strings.Replace(prices[0], ".00", "00", 1))
				log.Log(err)

				low, err := strconv.Atoi(strings.Replace(prices[0], ".00", "00", 1))
				log.Log(err)

				gorm = gorm.Where("JSON_EXTRACT(prices, \"$.US.final\") >= ?", high)
				gorm = gorm.Where("JSON_EXTRACT(prices, \"$.US.final\") <= ?", low)

			}

			// Score range
			scores := query.GetSearchSlice("scores")
			if len(scores) == 2 {

				gorm = gorm.Where("reviews_score >= ?", scores[0])
				gorm = gorm.Where("reviews_score <= ?", scores[1])

			}

			// Search
			search := query.GetSearchString("search")
			if search != "" {
				gorm = gorm.Where("name LIKE ?", "%"+search+"%")
			}

			// Count
			gorm.Count(&recordsFiltered)
			log.Log(gorm.Error)

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
			log.Log(gorm.Error)
		}

		wg.Done()
	}()

	// Get total
	var count int
	wg.Add(1)
	go func() {

		var err error
		count, err = db.CountApps()
		log.Log(err)

		wg.Done()
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

	response.output(w)
}
