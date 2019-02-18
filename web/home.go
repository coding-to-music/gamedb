package web

import (
	"net/http"
	"sync"

	"github.com/gamedb/website/db"
	"github.com/gamedb/website/log"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {

	t := homeTemplate{}
	t.Fill(w, r, "Home", "Stats and information on the Steam Catalogue.")

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.RanksCount, err = db.CountRanks()
		log.Err(err, r)

	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.AppsCount, err = db.CountApps()
		log.Err(err, r)

	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.PackagesCount, err = db.CountPackages()
		log.Err(err, r)

	}()

	wg.Wait()

	err := returnTemplate(w, r, "home", t)
	log.Err(err, r)
}

type homeTemplate struct {
	GlobalTemplate
	RanksCount    int
	AppsCount     int
	PackagesCount int
}
