package handlers

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gamedb/gamedb/cmd/frontend/helpers/session"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/i18n"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi/v5"
)

// Get prices ajax
func productPricesAjaxHandler(w http.ResponseWriter, r *http.Request, productType helpers.ProductType) {

	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return
	}

	// Get product
	var product helpers.ProductInterface

	if productType == helpers.ProductTypeApp {
		product, err = mongo.GetApp(id)
	} else {
		product, err = mongo.GetPackage(id)
	}
	if err != nil {
		log.ErrS(err)
		return
	}

	//
	var code = session.GetProductCC(r)

	// Get prices
	pricesResp, err := mongo.GetPricesForProduct(product.GetID(), product.GetProductType(), code)
	if err != nil {
		log.ErrS(err)
		return
	}

	// Make JSON response
	var response productPricesAjaxResponse
	response.Symbol = i18n.GetProdCC(code).Symbol

	for _, v := range pricesResp {
		response.Prices = append(response.Prices, []float64{float64(v.CreatedAt.Unix() * 1000), float64(v.PriceAfter) / 100})
	}

	// Add current price
	price := product.GetPrices().Get(code)
	if price.Exists {
		response.Prices = append(response.Prices, []float64{float64(time.Now().Unix()) * 1000, float64(price.Final) / 100})
	}

	// Convert a single dot into a line
	if len(response.Prices) == 1 {
		response.Prices = append(response.Prices, []float64{response.Prices[0][0] - 1000, response.Prices[0][1]})
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
