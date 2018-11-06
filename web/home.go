package web

import (
	"net/http"
	"sync"

	"github.com/gamedb/website/db"
	"github.com/gamedb/website/logging"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {

	t := homeTemplate{}
	t.Fill(w, r, "Home")

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {

		var err error
		t.RanksCount, err = db.CountRanks()
		logging.Error(err)

		wg.Done()
	}()

	wg.Add(1)
	go func() {

		var err error
		t.AppsCount, err = db.CountApps()
		logging.Error(err)

		wg.Done()
	}()

	wg.Add(1)
	go func() {

		var err error
		t.PackagesCount, err = db.CountPackages()
		logging.Error(err)

		wg.Done()
	}()

	wg.Wait()

	returnTemplate(w, r, "home", t)
}

type homeTemplate struct {
	GlobalTemplate
	RanksCount    int
	AppsCount     int
	PackagesCount int
}
