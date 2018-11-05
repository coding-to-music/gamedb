package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gamedb/website/db"
	"github.com/gamedb/website/logging"
	"github.com/gamedb/website/session"
	"github.com/go-chi/chi"
)

func PackageHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")

	idx, err := strconv.Atoi(id)
	if err != nil {
		returnErrorTemplate(w, r, 404, "Invalid package ID")
		return
	}

	// Get package
	pack, err := db.GetPackage(idx)
	if err != nil {

		if err == db.ErrNotFound {
			returnErrorTemplate(w, r, 404, "We can't find this package in our database, there may not be one with this ID.")
			return
		}

		logging.Error(err)
		returnErrorTemplate(w, r, 500, err.Error())
		return
	}

	// Redirect to correct slug
	if r.URL.Path != pack.GetPath() {
		http.Redirect(w, r, pack.GetPath(), 302)
		return
	}

	//
	var wg sync.WaitGroup

	var apps = map[int]db.App{}
	wg.Add(1)
	go func() {

		// Get apps
		appIDs, err := pack.GetAppIDs()
		logging.Error(err)

		for _, v := range appIDs {
			apps[v] = db.App{ID: v}
		}

		appRows, err := db.GetAppsByID(appIDs, []string{"id", "name", "icon", "type", "platforms", "dlc"})
		logging.Error(err)

		for _, v := range appRows {
			apps[v.ID] = v
		}

		wg.Done()
	}()

	var pricesString string
	var pricesCount int
	wg.Add(1)
	go func() {

		var code = session.GetCountryCode(r)

		// Get prices
		pricesResp, err := db.GetProductPrices(pack.ID, db.ProductTypePackage, code)
		logging.Error(err)

		pricesCount = len(pricesResp)

		var prices [][]float64

		for _, v := range pricesResp {
			prices = append(prices, []float64{float64(v.CreatedAt.Unix()), float64(v.PriceAfter) / 100})
		}

		// Add current price
		pricesStruct, err := pack.GetPrice(code)
		logging.Error(err)

		prices = append(prices, []float64{float64(time.Now().Unix()), float64(pricesStruct.Final) / 100})

		// Make into a JSON string
		pricesBytes, err := json.Marshal(prices)
		logging.Error(err)

		pricesString = string(pricesBytes)

		wg.Done()
	}()

	// Make banners
	banners := make(map[string][]string)
	var primary []string

	// if pack.GetExtended() == "prerelease" {
	// 	primary = append(primary, "This package is intended for developers and publishers only.")
	// }

	if len(primary) > 0 {
		banners["primary"] = primary
	}

	// Wait
	wg.Wait()

	// Template
	t := packageTemplate{}
	t.Fill(w, r, pack.GetName())
	t.Package = pack
	t.Apps = apps
	t.Prices = pricesString
	t.PricesCount = pricesCount

	returnTemplate(w, r, "package", t)
}

type packageTemplate struct {
	GlobalTemplate
	Package     db.Package
	Apps        map[int]db.App
	Banners     map[string][]string
	Prices      string
	PricesCount int
}
