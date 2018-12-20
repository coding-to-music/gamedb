package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/queue"
	"github.com/gamedb/website/session"
	"github.com/go-chi/chi"
)

func packageHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")

	idx, err := strconv.Atoi(id)
	if err != nil {
		returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "Invalid package ID.", Error: err})
		return
	}

	// Get package
	pack, err := db.GetPackage(idx)
	if err != nil {

		if err == db.ErrNotFound {
			returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "Sorry but we can not find this package."})
			return
		}

		returnErrorTemplate(w, r, errorTemplate{Code: 500, Message: "There was an issue retrieving the package.", Error: err})
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

		defer wg.Done()

		// Get apps
		appIDs, err := pack.GetAppIDs()
		log.Log(err)

		for _, v := range appIDs {
			apps[v] = db.App{ID: v}
		}

		appRows, err := db.GetAppsByID(appIDs, []string{"id", "name", "icon", "type", "platforms", "dlc"})
		log.Log(err)

		for _, v := range appRows {
			apps[v.ID] = v
		}

	}()

	var pricesString string
	wg.Add(1)
	go func() {

		defer wg.Done()

		var code = session.GetCountryCode(r)

		// Get prices
		pricesResp, err := db.GetProductPrices(pack.ID, db.ProductTypePackage, code)
		log.Log(err)

		var prices [][]float64
		for _, v := range pricesResp {
			prices = append(prices, []float64{float64(v.CreatedAt.Unix()), float64(v.PriceAfter) / 100})
		}

		// Add current price
		pricesStruct, err := pack.GetPrice(code)
		log.Log(err)

		prices = append(prices, []float64{float64(time.Now().Unix()), float64(pricesStruct.Final) / 100})

		// Make into a JSON string
		pricesBytes, err := json.Marshal(prices)
		log.Log(err)

		pricesString = string(pricesBytes)

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
	t.Fill(w, r, pack.GetName(), "")
	t.Package = pack
	t.Apps = apps
	t.Prices = pricesString

	// Update news, reviews etc
	func() {

		if helpers.IsBot(r.UserAgent()) {
			log.Info("Bots can't update packages")
			return
		}

		if pack.UpdatedAt.Unix() > time.Now().Add(time.Hour * -24).Unix() {
			log.Info("Too soon")
			return
		}

		err = queue.QueuePackage([]int{pack.ID})
		if err != nil {
			log.Log(err)
		} else {
			t.addToast(Toast{Title: "Update", Message: "Package has been queued for an update"})
		}
	}()

	// Get price
	t.Price = db.GetPriceFormatted(pack, session.GetCountryCode(r))

	err = returnTemplate(w, r, "package", t)
	log.Log(err)
}

type packageTemplate struct {
	GlobalTemplate
	Package db.Package
	Apps    map[int]db.App
	Banners map[string][]string
	Price   db.ProductPriceFormattedStruct
	Prices  string
}
