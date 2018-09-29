package web

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/steam-authority/steam-authority/db"
	"github.com/steam-authority/steam-authority/logger"
)

func PackagesHandler(w http.ResponseWriter, r *http.Request) {

	total, err := db.CountPackages()
	logger.Error(err)

	// Template
	t := packagesTemplate{}
	t.Fill(w, r, "Packages")
	t.Total = total

	returnTemplate(w, r, "packages", t)
}

type packagesTemplate struct {
	GlobalTemplate
	Total int
}

func PackagesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	query.FillFromURL(r.URL.Query())

	//
	var wg sync.WaitGroup

	// Get apps
	var packages []db.Package

	wg.Add(1)
	go func() {

		gorm, err := db.GetMySQLClient()
		if err != nil {

			logger.Error(err)

		} else {

			gorm = gorm.Model(&db.Package{})
			gorm = gorm.Select([]string{"id", "name", "billing_type", "license_type", "status", "apps", "updated_at"})

			gorm = query.SetOrderOffsetGorm(gorm, map[string]string{
				"0": "name",
				"1": "billing_type",
				"2": "license_type",
				"3": "status",
				"4": "apps",
				"5": "updated_at",
			})

			gorm = gorm.Limit(100)
			gorm = gorm.Find(&packages)

			logger.Error(gorm.Error)
		}

		wg.Done()
	}()

	// Get total
	var count int
	wg.Add(1)
	go func() {

		var err error
		count, err = db.CountPackages()
		logger.Error(err)

		wg.Done()
	}()

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.Itoa(count)
	response.RecordsFiltered = strconv.Itoa(count)
	response.Draw = query.Draw

	for _, v := range packages {

		apps, err := v.GetAppIDs()
		logger.Error(err)

		response.AddRow([]interface{}{
			v.ID,
			v.GetName(),
			v.GetBillingType(),
			v.GetLicenseType(),
			v.GetStatus(),
			len(apps),
			v.GetUpdatedNice(),
			v.GetPath(),
		})
	}

	response.output(w)
}
