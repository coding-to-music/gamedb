package pages

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/go-chi/chi"
	. "go.mongodb.org/mongo-driver/bson"
)

func SalesRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", salesHandler)
	r.Get("/sales.json", salesAjaxHandler)
	return r
}

func salesHandler(w http.ResponseWriter, r *http.Request) {

	t := salesTemplate{}
	t.addAssetChosen()
	t.addAssetSlider()
	t.addAssetCountdown()
	t.fill(w, r, "Offers", "")

	var wg sync.WaitGroup

	// Get tags
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Tags, err = sql.GetTagsForSelect()
		log.Err(err, r)
	}()

	// Count players
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Count, err = mongo.CountSales()
		log.Err(err, r)
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.HighestOrder, err = mongo.GetHighestSaleOrder()
		log.Err(err, r)
	}()

	// Get categories
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.Categories, err = sql.GetCategoriesForSelect()
		log.Err(err, r)
	}()

	// Upcoming days

	pst, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		log.Err(err, r)
	}

	upcomingSales := []upcomingSale{
		{time.Date(2019, 10, 28, 10, 0, 0, 0, pst), 4, "Halloween Sale", "🎃"},
		{time.Date(2019, 11, 26, 10, 0, 0, 0, pst), 7, "Autumn Sale", "🍁"},
		{time.Date(2019, 12, 19, 10, 0, 0, 0, pst), 14, "Winter Sale", "⛄"},
	}

	for _, v := range upcomingSales {
		if !v.Ended() {
			t.UpcomingSale = v
			break
		}
	}

	// Wait
	wg.Wait()

	returnTemplate(w, r, "sales", t)
}

type salesTemplate struct {
	GlobalTemplate
	Tags         []sql.Tag
	Categories   []sql.Category
	UpcomingSale upcomingSale
	HighestOrder int
	Count        int64
}

type upcomingSale struct {
	Start time.Time
	Days  int
	Name  string
	Icon  string
}

func (ud upcomingSale) ID() string {
	return "sale-" + strconv.FormatInt(ud.Start.Unix(), 10)
}

func (ud upcomingSale) Time() int64 {
	if ud.Start.Unix() < time.Now().Unix() {
		return ud.Start.AddDate(0, 0, ud.Days).Unix() * 1000
	} else {
		return ud.Start.Unix() * 1000
	}
}

func (ud upcomingSale) Ended() bool {
	return ud.Start.AddDate(0, 0, ud.Days).Unix() < time.Now().Unix()
}

func salesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	var filter = D{
		{"offer_end", M{"$gt": time.Now()}},
	}

	order := query.getSearchString("order")
	orderI, err := strconv.Atoi(strings.TrimSuffix(order, ".00"))
	if err == nil {
		filter = append(filter, E{Key: "sub_order", Value: M{"$lte": orderI - 1}})
	}

	//
	var wg sync.WaitGroup
	var code = helpers.GetProductCC(r)
	var offers []mongo.Sale

	// Get rows
	wg.Add(1)
	go func() {

		defer wg.Done()

		var columns = map[string]string{
			"0": "app_name",
			"1": "app_prices." + string(code),
			"2": "offer_percent",
			"3": "app_rating",
			"4": "offer_end",
			"5": "app_date",
		}

		var err error
		offers, err = mongo.GetAllSales(query.getOffset64(), 100, filter, query.getOrderMongo(columns, nil))
		if err != nil {
			log.Err(err, r)
			return
		}
	}()

	// Get count
	var count int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		count, err = mongo.CountDocuments(mongo.CollectionAppSales, nil, 0)
		if err != nil {
			log.Err(err, r)
		}
	}()

	// Get filtered count
	var filtered int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		filtered, err = mongo.CountDocuments(mongo.CollectionAppSales, filter, 0)
		if err != nil {
			log.Err(err, r)
		}
	}()

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = count
	response.RecordsFiltered = filtered
	response.Draw = query.Draw

	for _, offer := range offers {

		// Price
		var priceString string
		priceInt, ok := offer.AppPrices[code]
		if ok {
			cc := helpers.GetProdCC(code)
			priceString = helpers.FormatPrice(cc.CurrencyCode, priceInt)
		} else {
			priceString = "-"
		}

		// Lowest price
		var lowest bool
		lowestPriceInt, ok := offer.AppLowestPrice[code]
		if ok {
			lowest = priceInt <= lowestPriceInt
		}

		//
		response.AddRow([]interface{}{
			offer.AppID,          // 0
			offer.GetOfferName(), // 1
			offer.AppIcon,        // 2
			helpers.GetAppPath(offer.AppID, offer.AppName), // 3
			priceString,                           // 4
			offer.SalePercent,                     // 5
			math.Round(offer.AppRating*100) / 100, // 6
			offer.SaleEnd.String(),                // 7
			helpers.GetAppStoreLink(offer.AppID),  // 8
			offer.AppReleaseDate.String(),         // 9
			offer.GetType(),                       // 10
			lowest,                                // 11
		})
	}

	response.output(w, r)
}
