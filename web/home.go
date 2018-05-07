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
		logger.Error(err)

		wg.Done()

	}()

	var appsCount int
	wg.Add(1)
	go func() {

		appsCount, err = mysql.CountApps()
		logger.Error(err)

		wg.Done()

	}()

	var packagesCount int
	wg.Add(1)
	go func() {

		packagesCount, err = mysql.CountPackages()
		logger.Error(err)

		wg.Done()

	}()

	wg.Wait()

	t := homeTemplate{}
	t.Fill(w, r, "Home")

	t.RanksCount = ranksCount
	t.AppsCount = appsCount
	t.PackagesCount = packagesCount

	returnTemplate(w, r, "home", t)
}

type homeTemplate struct {
	GlobalTemplate
	RanksCount    int
	AppsCount     int
	PackagesCount int
}
