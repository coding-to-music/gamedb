package web

import (
	"math"
	"net/http"
	"strconv"
	"sync"

	"github.com/Jleagle/steam-go/steam"
	"github.com/Masterminds/squirrel"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/logging"
)

func AppsHandler(w http.ResponseWriter, r *http.Request) {

	// Template
	t := appsTemplate{}
	t.Fill(w, r, "Games")
	t.Types = db.GetTypesForSelect()

	//
	var err error
	var wg sync.WaitGroup

	// Get apps count
	wg.Add(1)
	go func() {

		t.Count, err = db.CountApps()
		logging.Error(err)

		wg.Done()

	}()

	// Get tags
	wg.Add(1)
	go func() {

		t.Tags, err = db.GetTagsForSelect()
		logging.Error(err)

		wg.Done()

	}()

	// Get genres
	wg.Add(1)
	go func() {

		t.Genres, err = db.GetGenresForSelect()
		logging.Error(err)

		wg.Done()

	}()

	// Get publishers
	wg.Add(1)
	go func() {

		t.Publishers, err = db.GetPublishersForSelect()
		logging.Error(err)

		wg.Done()

	}()

	// Get developers
	wg.Add(1)
	go func() {

		t.Developers, err = db.GetDevelopersForSelect()
		logging.Error(err)

		wg.Done()

	}()

	// Get most expensive app
	wg.Add(1)
	go func() {

		price, err := db.GetMostExpensiveApp(steam.CountryUS)
		logging.Error(err)

		// Convert cents to dollars
		t.ExpensiveApp = int(math.Ceil(float64(price) / 100))

		t.ExpensiveApp = 101 // todo, remove this line when apps have prices¬

		wg.Done()

	}()

	// Wait
	wg.Wait()

	returnTemplate(w, r, "apps", t)
}

type appsTemplate struct {
	GlobalTemplate
	Count        int
	ExpensiveApp int
	Types        map[string]string
	Tags         []db.Tag
	Genres       []db.Genre
	Publishers   []db.Publisher
	Developers   []db.Developer
}

func AppsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	query.FillFromURL(r.URL.Query())

	//
	var wg sync.WaitGroup

	// Get apps
	var apps []db.App
	var recordsFiltered int

	wg.Add(1)
	go func() {

		gorm, err := db.GetMySQLClient()
		if err != nil {

			logging.Error(err)

		} else {

			gorm = gorm.Model(db.App{})
			gorm = gorm.Select([]string{"id", "name", "icon", "reviews_score", "type", "dlc_count"})

			types := query.GetSearchSlice("types")
			if len(types) > 0 {
				gorm = gorm.Where("type IN (?)", types)
			}

			tags := query.GetSearchSlice("tags")
			if len(types) > 0 {

				var or squirrel.Or
				for _, v := range tags {
					or = append(or, squirrel.Eq{"JSON_CONTAINS(tags, '[\"" + v + "\"]')": 1})
				}
				sql, data, err := or.ToSql()
				logging.Error(err)

				gorm = gorm.Where(sql, data)
			}

			genres := query.GetSearchSlice("genres")
			if len(genres) > 0 {

				var or squirrel.Or
				for _, v := range genres {
					or = append(or, squirrel.Eq{"JSON_CONTAINS(genres, JSON_OBJECT('id', " + v + "))": 1})
				}
				sql, data, err := or.ToSql()
				logging.Error(err)

				gorm = gorm.Where(sql, data...)
			}

			developers := query.GetSearchSlice("developers")
			if len(developers) > 0 {

				var or squirrel.Or
				for _, v := range developers {
					or = append(or, squirrel.Eq{"JSON_CONTAINS(developers, '[\"" + v + "\"]')": 1})
				}
				sql, data, err := or.ToSql()
				logging.Error(err)

				gorm = gorm.Where(sql, data...)
			}

			publishers := query.GetSearchSlice("publishers")
			if len(publishers) > 0 {

				var or squirrel.Or
				for _, v := range publishers {
					or = append(or, squirrel.Eq{"JSON_CONTAINS(publishers, '[\"" + v + "\"]')": 1})
				}
				sql, data, err := or.ToSql()
				logging.Error(err)

				gorm = gorm.Where(sql, data...)
			}

			platforms := query.GetSearchSlice("platforms")
			if len(platforms) > 0 {

				var or squirrel.Or
				for _, v := range platforms {
					or = append(or, squirrel.Eq{"JSON_CONTAINS(platforms, '[\"" + v + "\"]')": 1})
				}
				sql, data, err := or.ToSql()
				logging.Error(err)

				gorm = gorm.Where(sql, data...)
			}

			prices := query.GetSearchSlice("prices")
			if len(prices) == 2 {

				// No apps have prices yet
				//gorm = gorm.Where("JSON_EXTRACT(prices, \"$.US.final\") >= ?", prices[0]+"00")
				//gorm = gorm.Where("JSON_EXTRACT(prices, \"$.US.final\") <= ?", prices[1]+"00")

			}

			scores := query.GetSearchSlice("scores")
			if len(prices) == 2 {

				gorm = gorm.Where("reviews_score >= ?", scores[0])
				gorm = gorm.Where("reviews_score <= ?", scores[1])

			}

			// Get count, must be above setting limit
			gorm.Count(&recordsFiltered)
			logging.Error(gorm.Error)

			//
			gorm = gorm.Limit(100)
			gorm = query.SetOrderOffsetGorm(gorm, steam.CountryUS, map[string]string{
				"0": "name",
				"2": "reviews_score",
				"3": "dlc_count",
				"4": "price",
			})

			// Get rows
			gorm = gorm.Find(&apps)
			logging.Error(gorm.Error)
		}

		wg.Done()
	}()

	// Get total
	var count int
	wg.Add(1)
	go func() {

		var err error
		count, err = db.CountApps()
		logging.Error(err)

		wg.Done()
	}()

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.Itoa(count)
	response.RecordsFiltered = strconv.Itoa(recordsFiltered)
	response.Draw = query.Draw

	for _, v := range apps {
		response.AddRow(v.OutputForJSON())
	}

	response.output(w)
}
