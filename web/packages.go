package web

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
)

const (
	PackagesLimit = 100
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
	packages, err := mysql.GetLatestPackages(PackagesLimit, page)
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
		limit: PackagesLimit,
		total: total,
	}

	returnTemplate(w, r, "packages", template)
}

type packagesTemplate struct {
	GlobalTemplate
	Packages   []mysql.Package
	Pagination Pagination
}

func PackageHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		returnErrorTemplate(w, r, 404, "Invalid package ID")
		return
	}

	pack, err := mysql.GetPackage(id)
	if err != nil {

		if err == mysql.ErrNotFound {
			returnErrorTemplate(w, r, 404, "We can't find this package in our database, there may not be one with this ID.")
			return
		}

		logger.Error(err)
		returnErrorTemplate(w, r, 500, err.Error())
		return
	}

	appIDs, err := pack.GetApps()
	if err != nil {
		logger.Error(err)
	}

	apps, err := mysql.GetApps(appIDs, []string{"id", "icon", "type", "platforms", "dlc"})
	if err != nil {
		logger.Error(err)
	}
	// Make banners
	banners := make(map[string][]string)
	var primary []string

	// if pack.GetExtended() == "prerelease" {
	// 	primary = append(primary, "This package is intended for developers and publishers only.")
	// }

	if len(primary) > 0 {
		banners["primary"] = primary
	}

	// Template
	template := packageTemplate{}
	template.Fill(r, pack.GetName())
	template.Package = pack
	template.Apps = apps
	template.ExtendedKeys = mysql.PackageExtendedKeys
	template.ControllerKeys = mysql.PackageControllerKeys

	returnTemplate(w, r, "package", template)
}

type packageTemplate struct {
	GlobalTemplate
	Package        mysql.Package
	Apps           []mysql.App
	ExtendedKeys   map[string]string
	ControllerKeys map[string]string
	Banners        map[string][]string
}
