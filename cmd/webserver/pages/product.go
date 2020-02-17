package pages

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
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
	var product helpers.ProductInterface

	if productType == helpers.ProductTypeApp {
		product, err = mongo.GetApp(idx)
	} else {
		product, err = mongo.GetPackage(idx)
	}
	if err != nil {
		log.Err(err, r)
		return
	}

	// Get code
	code := steamapi.ProductCC(r.URL.Query().Get("code"))
	if code == "" || !helpers.IsValidProdCC(code) {
		code = helpers.GetProductCC(r)
	}

	// Get prices
	pricesResp, err := mongo.GetPricesForProduct(product.GetID(), product.GetProductType(), code)
	if err != nil {
		log.Err(err, r)
		return
	}

	// Make JSON response
	var response productPricesAjaxResponse
	response.Symbol = helpers.GetProdCC(code).Symbol

	for _, v := range pricesResp {
		response.Prices = append(response.Prices, []float64{float64(v.CreatedAt.Unix() * 1000), float64(v.PriceAfter) / 100})
	}

	// Add current price
	price := product.GetPrices().Get(code)
	if price.Exists {
		response.Prices = append(response.Prices, []float64{float64(time.Now().Unix()) * 1000, float64(price.Final) / 100})
	}

	// Sort prices for Highcharts
	sort.Slice(response.Prices, func(i, j int) bool {
		return response.Prices[i][0] < response.Prices[j][0]
	})

	// Return
	returnJSON(w, r, response)
}

type productPricesAjaxResponse struct {
	Prices [][]float64 `json:"prices"`
	Symbol string      `json:"symbol"`
}
