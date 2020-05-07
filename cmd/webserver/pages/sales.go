package pages

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/datatable"
	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
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
	t.fill(w, r, "Offers", "Discounted games")

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

		var code = session.GetProductCC(r)

		var err error
		t.AppTypes, err = mongo.GetAppsGroupedByType(code)
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
		{time.Date(2019, 11, 10, 8, 0, 0, 0, pst), 2, "Singles' Day", ""},
		{time.Date(2019, 11, 26, 10, 0, 0, 0, pst), 7, "Autumn Sale", "🍁"},
		{time.Date(2019, 12, 12, 10, 0, 0, 0, pst), 1, "Game Awards", "🕹️"},
		{time.Date(2019, 12, 19, 10, 0, 0, 0, pst), 14, "Winter Sale", "⛄"},
		{time.Date(2020, 01, 23, 10, 0, 0, 0, pst), 4, "Lunar New Year Sale", "🌑"},
	}

	for _, v := range upcomingSales {
		if !v.Ended() {
			t.UpcomingSale = v
			break
		}
	}

	// Wait
	wg.Wait()

	t.SaleTypes, err = mongo.GetUniqueSaleTypes()
	log.Err(err, r)

	returnTemplate(w, r, "sales", t)
}

type salesTemplate struct {
	GlobalTemplate
	Tags         []sql.Tag
	Categories   []sql.Category
	UpcomingSale upcomingSale
	HighestOrder int
	AppTypes     []mongo.AppTypeCount
	SaleTypes    []string
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
	if ud.Start.Before(time.Now()) {
		return ud.Start.AddDate(0, 0, ud.Days).Unix() * 1000
	} else {
		return ud.Start.Unix() * 1000
	}
}

func (ud upcomingSale) Started() bool {
	return ud.Start.Before(time.Now())
}

func (ud upcomingSale) Ended() bool {
	return ud.Start.AddDate(0, 0, ud.Days).Before(time.Now())
}

func (ud upcomingSale) Show() bool {
	return ud.Name != "" && (ud.Time() < time.Now().AddDate(0, 0, 7).Unix())
}

func salesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := datatable.NewDataTableQuery(r, false)

	var code = session.GetProductCC(r)
	var countLock sync.Mutex
	var baseFilter = bson.D{
		{Key: "offer_end", Value: bson.M{"$gt": time.Now()}},
	}
	var filter = baseFilter

	search := helpers.RegexNonAlphaNumericSpace.ReplaceAllString(query.GetSearchString("search"), "")
	if search != "" {

		quoted := regexp.QuoteMeta(search)

		filter = append(filter, bson.E{Key: "$or", Value: bson.A{
			bson.M{"app_name": bson.M{"$regex": quoted, "$options": "i"}},
			bson.M{"offer_name": bson.M{"$regex": quoted, "$options": "i"}},
		}})
	}

	// Index
	index := query.GetSearchString("index")
	if index != "" {
		orderI, err := strconv.Atoi(strings.TrimSuffix(index, ".00"))
		if err == nil {
			filter = append(filter, bson.E{Key: "sub_order", Value: bson.M{"$lte": orderI - 1}})
		}
	}

	// Score
	scores := query.GetSearchSlice("score")
	if len(scores) == 2 {

		low, err := strconv.Atoi(strings.TrimSuffix(scores[0], ".00"))
		log.Err(err, r)

		high, err := strconv.Atoi(strings.TrimSuffix(scores[1], ".00"))
		log.Err(err, r)

		if low > 0 {
			filter = append(filter, bson.E{Key: "app_rating", Value: bson.M{"$gte": low}})
		}
		if high < 100 {
			filter = append(filter, bson.E{Key: "app_rating", Value: bson.M{"$lte": high}})
		}
	}

	// Price
	prices := query.GetSearchSlice("price")
	if len(prices) == 2 {

		low, err := strconv.Atoi(strings.TrimSuffix(prices[0], ".00"))
		log.Err(err, r)

		high, err := strconv.Atoi(strings.TrimSuffix(prices[1], ".00"))
		log.Err(err, r)

		if low > 0 {
			filter = append(filter, bson.E{Key: "app_prices." + string(code), Value: bson.M{"$gte": low * 100}})
		}
		if high < 100 {
			filter = append(filter, bson.E{Key: "app_prices." + string(code), Value: bson.M{"$lte": high * 100}})
		}
	}

	// Discount
	discounts := query.GetSearchSlice("discount")
	if len(discounts) == 2 {

		low, err := strconv.Atoi(strings.TrimSuffix(discounts[0], ".00"))
		log.Err(err, r)

		high, err := strconv.Atoi(strings.TrimSuffix(discounts[1], ".00"))
		log.Err(err, r)

		if low > 0 {
			filter = append(filter, bson.E{Key: "offer_percent", Value: bson.M{"$lte": -low}})
		}
		if high < 100 {
			filter = append(filter, bson.E{Key: "offer_percent", Value: bson.M{"$gte": -high}})
		}
	}

	// App type
	appTypes := query.GetSearchSlice("app-type")
	if len(appTypes) > 0 {

		var or bson.A
		for _, v := range appTypes {
			or = append(or, bson.M{"app_type": v})
		}
		filter = append(filter, bson.E{Key: "$or", Value: or})
	}

	// Sale type
	saleTypes := query.GetSearchSlice("sale-type")
	if len(saleTypes) > 0 {

		var or bson.A
		for _, v := range saleTypes {
			or = append(or, bson.M{"offer_type": v})
		}
		filter = append(filter, bson.E{Key: "$or", Value: or})
	}

	// Tag in
	tagsIn := query.GetSearchSlice("tags-in")
	if len(tagsIn) > 0 {

		var or bson.A
		for _, tag := range tagsIn {
			i, err := strconv.Atoi(tag)
			if err == nil {
				or = append(or, bson.M{"app_tags": i})
			}
		}
		filter = append(filter, bson.E{Key: "$or", Value: or})
	}

	// Tag out
	tagsOut := query.GetSearchSlice("tags-out")
	if len(tagsOut) > 0 {

		var or bson.A
		for _, tag := range tagsOut {
			i, err := strconv.Atoi(tag)
			if err == nil {
				or = append(or, bson.M{"app_tags": bson.M{"$ne": i}})
			}
		}
		filter = append(filter, bson.E{Key: "$or", Value: or})
	}

	// Categories
	categories := query.GetSearchSlice("categories")
	if len(categories) > 0 {

		var in bson.A
		for _, tag := range categories {
			i, err := strconv.Atoi(tag)
			if err == nil {
				in = append(in, i)
			}
		}
		filter = append(filter, bson.E{Key: "app_categories", Value: bson.M{"$in": in}})
	}

	// Platforms
	platforms := query.GetSearchSlice("platforms")
	if len(platforms) > 0 {

		var in bson.A
		for _, tag := range platforms {
			in = append(in, tag)
		}
		filter = append(filter, bson.E{Key: "app_platforms", Value: bson.M{"$in": in}})
	}

	//
	var wg sync.WaitGroup
	var offers []mongo.Sale

	// Get rows
	wg.Add(1)
	go func() {

		defer wg.Done()

		var columns = map[string]string{
			"0": "offer_name",
			"1": "app_prices." + string(code),
			"2": "offer_percent",
			"3": "app_rating",
			"4": "offer_end",
			"5": "app_date",
		}

		order := query.GetOrderMongo(columns)
		order = append(order, bson.E{Key: "app_rating", Value: -1})
		order = append(order, bson.E{Key: "app_name", Value: 1})
		order = append(order, bson.E{Key: "sub_order", Value: 1})

		var err error
		offers, err = mongo.GetAllSales(query.GetOffset64(), 100, filter, order)
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
		countLock.Lock()
		count, err = mongo.CountDocuments(mongo.CollectionAppSales, baseFilter, 0)
		countLock.Unlock()
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
		countLock.Lock()
		filtered, err = mongo.CountDocuments(mongo.CollectionAppSales, filter, 0)
		countLock.Unlock()
		if err != nil {
			log.Err(err, r)
		}
	}()

	// Wait
	wg.Wait()

	var response = datatable.NewDataTablesResponse(r, query, count, filtered)
	for _, offer := range offers {

		response.AddRow([]interface{}{
			offer.AppID,          // 0
			offer.GetOfferName(), // 1
			offer.AppIcon,        // 2
			helpers.GetAppPath(offer.AppID, offer.AppName), // 3
			offer.GetPriceString(code),                     // 4
			offer.SalePercent,                              // 5
			offer.GetAppRating(),                           // 6
			offer.SaleEnd.String(),                         // 7
			helpers.GetAppStoreLink(offer.AppID),           // 8
			offer.AppReleaseDate.String(),                  // 9
			offer.GetType(),                                // 10
			offer.IsLowest(code),                           // 11
			offer.SaleEndEstimate,                          // 12
			helpers.GetAppType(offer.AppType),              // 13
			offer.AppReleaseDateString,                     // 14
			offer.AppCategories,                            // 15
		})
	}

	returnJSON(w, r, response)
}
