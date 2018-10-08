package web

import (
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/steam-authority/steam-authority/db"
	"github.com/steam-authority/steam-authority/logger"
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
		logger.Error(err)

		wg.Done()

	}()

	// Get apps count
	var count int
	wg.Add(1)
	go func() {

		count, err = db.CountApps()
		logger.Error(err)

		wg.Done()

	}()

	// Wait
	wg.Wait()

	// Make pagination path
	values := r.URL.Query()
	values.Del("p")

	path := "/games?" + values.Encode() + "&p="
	path = strings.Replace(path, "?&", "?", 1)

	// Template
	t := appsTemplate{}
	t.Fill(w, r, "Games")
	t.Apps = apps
	t.Count = count

	returnTemplate(w, r, "apps", t)
}

type appsTemplate struct {
	GlobalTemplate
	Apps  []db.App
	Count int
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

			logger.Error(err)

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

			logger.Error(gorm.Error)
		}

		wg.Done()
	}()

	// Get total
	var count int
	wg.Add(1)
	go func() {

		var err error
		count, err = db.CountApps()
		logger.Error(err)

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
