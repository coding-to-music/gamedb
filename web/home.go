package web

import (
	"net/http"
	"sync"

	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {

	var wg sync.WaitGroup
	var err error

	var ranksCount int
	wg.Add(1)
	go func() {

		ranksCount, err = datastore.CountRanks()
		if err != nil {
			logger.Error(err)
		}

		wg.Done()

	}()

	var appsCount int
	wg.Add(1)
	go func() {

		appsCount, err = mysql.CountApps()
		if err != nil {
			logger.Error(err)
		}

		wg.Done()

	}()

	var packagesCount int
	wg.Add(1)
	go func() {

		packagesCount, err = mysql.CountPackages()
		if err != nil {
			logger.Error(err)
		}

		wg.Done()

	}()

	wg.Wait()

	template := homeTemplate{}
	template.Fill(r, "Home")

	template.RanksCount = ranksCount
	template.AppsCount = appsCount
	template.PackagesCount = packagesCount

	returnTemplate(w, r, "home", template)
}

type homeTemplate struct {
	GlobalTemplate
	RanksCount    int
	AppsCount     int
	PackagesCount int
}
