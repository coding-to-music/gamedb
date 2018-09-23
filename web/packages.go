package web

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/steam-authority/steam-authority/db"
	"github.com/steam-authority/steam-authority/logger"
)

const (
	packagesLimit = 100
)

func PackagesHandler(w http.ResponseWriter, r *http.Request) {

	// Get page number
	page, err := strconv.Atoi(r.URL.Query().Get("p"))
	if err != nil {
		page = 1
	}

	var wg sync.WaitGroup

	// Get total changes
	var total int
	wg.Add(1)
	go func() {

		total, err = db.CountPackages()
		logger.Error(err)

		wg.Done()

	}()

	// Get changes
	var packages []db.Package
	wg.Add(1)
	go func() {

		packages, err = db.GetLatestPackages(packagesLimit, page)
		logger.Error(err)

		wg.Done()

	}()

	// Wait
	wg.Wait()

	// Template
	t := packagesTemplate{}
	t.Fill(w, r, "Packages")
	t.Packages = packages
	t.Total = total
	t.Pagination = Pagination{
		path:  "/packages?p=",
		page:  page,
		limit: packagesLimit,
		total: total,
	}

	returnTemplate(w, r, "packages", t)
}

type packagesTemplate struct {
	GlobalTemplate
	Packages   []db.Package
	Pagination Pagination
	Total      int
}
