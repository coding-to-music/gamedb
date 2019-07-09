package pages

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/go-chi/chi"
)

// Get prices ajax
func productPricesAjaxHandler(w http.ResponseWriter, r *http.Request, productType helpers.ProductType) {

	id := chi.URLParam(r, "id")
	if id == "" {
		log.Err("invalid id", r)
		return
	}

	idx, err := strconv.Atoi(id)
	if err != nil {
		log.Err("invalid id", r)
		return
	}

	// Get product
	var product sql.ProductInterface

	if productType == helpers.ProductTypeApp {
		product, err = sql.GetApp(idx, nil)
	} else {
		product, err = sql.GetPackage(idx)
	}
	if err != nil {
		log.Err(err, r)
		return
	}

	// Get code
	code := steam.CountryCode(r.URL.Query().Get("code"))
	if code == "" {
		code = helpers.GetCountryCode(r)
	}

	if code == "" {
		log.Err("no code given", r)
		return
	}

	// Get prices
	pricesResp, err := mongo.GetPricesForProduct(product.GetID(), product.GetProductType(), code)
	if err != nil {
		log.Err(err, r)
		return
	}

	// Get locale
	locale, err := helpers.GetLocaleFromCountry(code)
	if err != nil {
		log.Err(err, r)
		return
	}

	// Make JSON response
	var response productPricesAjaxResponse
	response.Symbol = locale.CurrencySymbol

	for _, v := range pricesResp {
		response.Prices = append(response.Prices, []float64{float64(v.CreatedAt.Unix() * 1000), float64(v.PriceAfter) / 100})
	}

	// Add current price
	price, err := product.GetPrice(code)
	err = helpers.IgnoreErrors(err, sql.ErrMissingCountryCode)
	if err != nil {
		log.Err(err, r)
		return
	}

	response.Prices = append(response.Prices, []float64{float64(time.Now().Unix()) * 1000, float64(price.Final) / 100})

	// Sort prices for Highcharts
	sort.Slice(response.Prices, func(i, j int) bool {
		return response.Prices[i][0] < response.Prices[j][0]
	})

	// Return
	err = returnJSON(w, r, response)
	log.Err(err, r)
}

type productPricesAjaxResponse struct {
	Prices [][]float64 `json:"prices"`
	Symbol string      `json:"symbol"`
}
