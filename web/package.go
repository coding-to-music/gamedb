package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi"
	slugify "github.com/gosimple/slug"
	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
)

func PackageHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	slug := chi.URLParam(r, "slug")

	idx, err := strconv.Atoi(id)
	if err != nil {
		returnErrorTemplate(w, r, 404, "Invalid package ID")
		return
	}

	// Get package
	pack, err := mysql.GetPackage(idx)
	if err != nil {

		if err == mysql.ErrNotFound {
			returnErrorTemplate(w, r, 404, "We can't find this package in our database, there may not be one with this ID.")
			return
		}

		logger.Error(err)
		returnErrorTemplate(w, r, 500, err.Error())
		return
	}

	// Redirect to correct slug
	correctSLug := slugify.Make(pack.GetName())
	if slug != correctSLug {
		http.Redirect(w, r, "/packages/"+id+"/"+correctSLug, 302)
		return
	}

	//
	var wg sync.WaitGroup

	var apps []mysql.App
	wg.Add(1)
	go func() {

		// Get apps
		appIDs, err := pack.GetApps()
		if err != nil {
			logger.Error(err)
		}

		apps, err = mysql.GetApps(appIDs, []string{"id", "icon", "type", "platforms", "dlc"})
		if err != nil {
			logger.Error(err)
		}

		wg.Done()
	}()

	var pricesString string
	var pricesCount int
	wg.Add(1)
	go func() {

		// Get prices
		pricesResp, err := datastore.GetPackagePrices(pack.ID)
		if err != nil {
			logger.Error(err)
		}

		pricesCount = len(pricesResp)

		var prices [][]float64

		for _, v := range pricesResp {

			prices = append(prices, []float64{float64(v.CreatedAt.Unix()), float64(v.PriceFinal) / 100})
		}

		// Add current price
		prices = append(prices, []float64{float64(time.Now().Unix()), float64(pack.PriceFinal) / 100})

		// Make into a JSON string
		pricesBytes, err := json.Marshal(prices)
		if err != nil {
			logger.Error(err)
		}

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
	t.Fill(r, pack.GetName())
	t.Package = pack
	t.Apps = apps
	t.Prices = pricesString
	t.PricesCount = pricesCount

	returnTemplate(w, r, "package", t)
}

type packageTemplate struct {
	GlobalTemplate
	Package     mysql.Package
	Apps        []mysql.App
	Banners     map[string][]string
	Prices      string
	PricesCount int
}
