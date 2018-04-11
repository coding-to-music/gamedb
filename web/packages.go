package web

import (
	"net/http"
	"strconv"

	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
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

	// Get total
	total, err := mysql.CountPackages()
	if err != nil {
		logger.Error(err)
	}

	// Get packages
	packages, err := mysql.GetLatestPackages(packagesLimit, page)
	if err != nil {
		logger.Error(err)
	}

	// Template
	template := packagesTemplate{}
	template.Fill(r, "Packages")
	template.Packages = packages
	template.Pagination = Pagination{
		path:  "/packages?p=",
		page:  page,
		limit: packagesLimit,
		total: total,
	}

	returnTemplate(w, r, "packages", template)
}

type packagesTemplate struct {
	GlobalTemplate
	Packages   []mysql.Package
	Pagination Pagination
}
