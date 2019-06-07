package pages

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/Jleagle/steam-go/steam"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/go-chi/chi"
)

func AppsRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", appsHandler)
	r.Get("/apps.json", appsAjaxHandler)
	r.Mount("/{id}", AppRouter())
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

		var err error
		t.Count, err = sql.CountApps()
		t.Description = "A live database of " + template.HTML(humanize.Comma(int64(t.Count))) + " Steam games."
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

	}()

	// Wait
	wg.Wait()

	// t.Columns = allColumns

	err := returnTemplate(w, r, "apps", t)
	log.Err(err, r)
}

type appsTemplate struct {
	GlobalTemplate
	Count      int
	Types      []sql.AppType
	Tags       []sql.Tag
	Genres     []sql.Genre
	Publishers []sql.Publisher
	Developers []sql.Developer
	Columns    []TableColumn
}

type TableColumn struct {
	Name    string
	Columns []string
}

// var (
// 	allColumns = map[string]TableColumn{
// 		"name":    {Name: "Name", Columns: []string{"id", "name", "icon"}},
// 		"type":    {Name: "Type", Columns: []string{"type"}},
// 		"score":   {Name: "Score", Columns: []string{"score"}},
// 		"price":   {Name: "Price", Columns: []string{"price"}},
// 		"updated": {Name: "Updated At", Columns: []string{"updated"}},
// 	}
//
// 	defaultColumns = []string{"name", "type", "score", "price", "updated"}
// )

func appsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	// Get columns
	// if len(query.Columns) == 0 {
	// 	query.Columns = defaultColumns
	// }
	//
	// var columns = []string{"id"}
	// for _, v := range query.Columns {
	//
	// 	_, ok := allColumns[v]
	// 	if ok {
	// 		columns = append(columns, allColumns[v].Columns...)
	// 	}
	// }

	//
	var code = getCountryCode(r)
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
		gorm = gorm.Select([]string{"id", "name", "icon", "type", "reviews_score", "prices", "player_peak_week"})

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
		prices := query.getSearchSlice("prices")
		if len(prices) == 2 {

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
			if high < 100*100 {
				gorm = gorm.Where("COALESCE("+column+", 0) <= ?", high)
			}

		}

		// Score range
		scores := query.getSearchSlice("scores")
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
		search := query.getSearchString("search")
		if search != "" {
			gorm = gorm.Where("name LIKE ?", "%"+search+"%")
		}

		// Count
		gorm = gorm.Count(&recordsFiltered)
		log.Err(gorm.Error)

		// Order, offset, limit
		gorm = gorm.Limit(100)
		gorm = query.setOrderOffsetGorm(gorm, code, map[string]string{
			"0": "name",
			"2": "player_peak_week",
			"3": "reviews_score",
			"4": "price",
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
		count, err = sql.CountApps()
		log.Err(err, r)

	}()

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = int64(count)
	response.RecordsFiltered = recordsFiltered
	response.Draw = query.Draw

	for _, v := range apps {
		response.AddRow(v.OutputForJSON(code))
	}

	response.output(w, r)
}
