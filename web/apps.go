package web

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/steam-authority/steam-authority/db"
	"github.com/steam-authority/steam-authority/logging"
)

const (
	appsSearchLimit = 100
)

func AppsHandler(w http.ResponseWriter, r *http.Request) {

	// Get page number
	page, err := strconv.Atoi(r.URL.Query().Get("p"))
	if err != nil {
		page = 1
	}

	var wg sync.WaitGroup

	// Get apps
	var apps []db.App
	wg.Add(1)
	go func() {

		apps, err = db.SearchApps(r.URL.Query(), appsSearchLimit, page, "id DESC", []string{})
		logging.Error(err)

		wg.Done()

	}()

	// Get apps count
	var count int
	wg.Add(1)
	go func() {

		count, err = db.CountApps()
		logging.Error(err)

		wg.Done()

	}()

	// Get tags
	var tags []db.Tag
	wg.Add(1)
	go func() {

		tags, err = db.GetTagsForSelect()
		logging.Error(err)

		wg.Done()

	}()

	// Get genres
	var genres []db.Genre
	wg.Add(1)
	go func() {

		genres, err = db.GetGenresForSelect()
		logging.Error(err)

		wg.Done()

	}()

	// Get genres
	var publishers []db.Publisher
	wg.Add(1)
	go func() {

		publishers, err = db.GetPublishersForSelect()
		logging.Error(err)

		wg.Done()

	}()

	// Get genres
	var developers []db.Developer
	wg.Add(1)
	go func() {

		developers, err = db.GetDevelopersForSelect()
		logging.Error(err)

		wg.Done()

	}()

	// Wait
	wg.Wait()

	// Make pagination path
	values := r.URL.Query()
	values.Del("p")

	// Template
	t := appsTemplate{}
	t.Fill(w, r, "Games")
	t.Count = count
	t.Apps = apps
	t.Tags = tags
	t.Genres = genres
	t.Publishers = publishers
	t.Developers = developers
	t.Types = db.GetTypesForSelect()

	returnTemplate(w, r, "apps", t)
}

type appsTemplate struct {
	GlobalTemplate
	Count      int
	Types      map[string]string
	Apps       []db.App
	Tags       []db.Tag
	Genres     []db.Genre
	Publishers []db.Publisher
	Developers []db.Developer
}

func AppsAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	query.FillFromURL(r.URL.Query())

	//
	var wg sync.WaitGroup

	// Get apps
	var apps []db.App

	wg.Add(1)
	go func() {

		gorm, err := db.GetMySQLClient()
		if err != nil {

			logging.Error(err)

		} else {

			//gorm = gorm.Model(&db.App{})
			gorm = gorm.Select([]string{"id", "name", "icon", "reviews_score", "type", "dlc_count"})

			gorm = query.SetOrderOffsetGorm(gorm, map[string]string{
				"0": "name",
				"2": "reviews_score",
				"3": "dlc_count",
			})

			gorm = gorm.Limit(100)
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
	response.RecordsFiltered = strconv.Itoa(count)
	response.Draw = query.Draw

	for _, v := range apps {
		response.AddRow(v.OutputForJSON())
	}

	response.output(w)
}
