package pages

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gamedb/gamedb/cmd/webserver/session"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/go-chi/chi"
)

func PackageRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", packageHandler)
	r.Get("/prices.json", packagePricesAjaxHandler)
	r.Get("/{slug}", packageHandler)
	return r
}

func packageHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	setCacheHeaders(w, time.Hour)

	id := chi.URLParam(r, "id")

	idx, err := strconv.Atoi(id)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "Invalid package ID.", Error: err})
		return
	}

	// Get package
	pack, err := sql.GetPackage(idx, []string{})
	if err != nil {

		if err == sql.ErrRecordNotFound {
			returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "Sorry but we can not find this package."})
			return
		}

		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the package.", Error: err})
		return
	}

	//
	var wg sync.WaitGroup

	var apps = map[int]sql.App{}
	wg.Add(1)
	go func() {

		defer wg.Done()

		// Get apps
		appIDs, err := pack.GetAppIDs()
		if err != nil {
			log.Err(err, r)
			return
		}

		for _, v := range appIDs {
			apps[v] = sql.App{ID: v}
		}

		appRows, err := sql.GetAppsByID(appIDs, []string{"id", "name", "icon", "type", "platforms", "dlc"})
		if err != nil {
			log.Err(err, r)
			return
		}

		for _, v := range appRows {
			apps[v.ID] = v
		}

	}()

	var bundles []sql.Bundle
	wg.Add(1)
	go func() {

		defer wg.Done()

		gorm, err := sql.GetMySQLClient()
		if err != nil {
			log.Err(err, r)
			return
		}

		gorm = gorm.Where("JSON_CONTAINS(package_ids, '[" + strconv.Itoa(pack.ID) + "]')")
		gorm = gorm.Find(&bundles)
		if gorm.Error != nil {
			log.Err(gorm.Error, r)
			return
		}
	}()

	// Wait
	wg.Wait()

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
	t := packageTemplate{}
	t.fill(w, r, pack.GetName(), "")
	t.metaImage = pack.GetMetaImage()
	t.addAssetHighCharts()
	t.Package = pack
	t.Apps = apps
	t.Bundles = bundles

	// Update news, reviews etc
	func() {

		if helpers.IsBot(r.UserAgent()) {
			return
		}

		if pack.UpdatedAt.Unix() > time.Now().Add(time.Hour * -24).Unix() {
			return
		}

		err = queue.ProducePackage(pack.ID)
		if err != nil {
			log.Err(err, r)
		} else {
			t.addToast(Toast{Title: "Update", Message: "Package has been queued for an update"})
		}
	}()

	// Get price
	t.Price = sql.GetPriceFormatted(pack, session.GetCountryCode(r))

	t.Prices, err = t.Package.GetPrices()
	log.Err(err)

	t.Extended, err = pack.GetExtended()
	log.Err(err)

	t.Controller, err = pack.GetController()
	log.Err(err)

	t.DepotIDs, err = pack.GetDepotIDs()
	log.Err(err)

	err = returnTemplate(w, r, "package", t)
	log.Err(err, r)
}

type packageTemplate struct {
	GlobalTemplate
	Apps       map[int]sql.App
	Bundles    []sql.Bundle
	Banners    map[string][]string
	Controller sql.PICSController
	DepotIDs   []int
	Extended   sql.PICSExtended
	Package    sql.Package
	Price      sql.ProductPriceFormattedStruct
	Prices     sql.ProductPrices
}

func (p packageTemplate) ShowDev() bool {

	return len(p.Extended) > 0 || len(p.Controller) > 0 || len(p.DepotIDs) > 0
}

func packagePricesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{"code"})
	if ret {
		return
	}

	setCacheHeaders(w, time.Hour*3)

	productPricesAjaxHandler(w, r, helpers.ProductTypePackage)
}
