package web

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/logging"
)

func PackagesHandler(w http.ResponseWriter, r *http.Request) {

	total, err := db.CountPackages()
	logging.Error(err)

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

			logging.Error(err)

		} else {

			gorm = gorm.Model(&db.Package{})
			gorm = gorm.Select([]string{"id", "name", "billing_type", "license_type", "status", "apps_count", "updated_at"})

			gorm = query.SetOrderOffsetGorm(gorm, steam.CountryUS, map[string]string{
				"0": "name",
				"4": "apps_count",
				"5": "updated_at",
			})

			gorm = gorm.Limit(100)
			gorm = gorm.Find(&packages)

			logging.Error(gorm.Error)
		}

		wg.Done()
	}()

	// Get total
	var count int
	wg.Add(1)
	go func() {

		var err error
		count, err = db.CountPackages()
		logging.Error(err)

		wg.Done()
	}()

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.Itoa(count)
	response.RecordsFiltered = strconv.Itoa(count)
	response.Draw = query.Draw

	for _, v := range packages {

		response.AddRow([]interface{}{
			v.ID,
			v.GetName(),
			v.GetBillingType(),
			v.GetLicenseType(),
			v.GetStatus(),
			v.AppsCount,
			v.GetUpdatedUnix(),
			v.GetUpdatedNice(),
			v.GetPath(),
		})
	}

	response.output(w)
}
