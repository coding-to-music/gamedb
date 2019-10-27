package pages

import (
	"net/http"
	"strconv"
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
		{time.Date(2019, 10, 28, 10, 0, 0, 0, pst), time.Date(2019, 11, 1, 10, 0, 0, 0, pst), "Halloween Sale", "🎃"},
		{time.Date(2019, 11, 26, 10, 0, 0, 0, pst), time.Date(2019, 12, 3, 10, 0, 0, 0, pst), "Autumn Sale", "🍁"},
		{time.Date(2019, 12, 19, 10, 0, 0, 0, pst), time.Date(2019, 01, 2, 10, 0, 0, 0, pst), "Winter Sale", "⛄"},
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
}

type upcomingSale struct {
	Start time.Time
	End   time.Time
	Name  string
	Icon  string
}

func (ud upcomingSale) ID() string {
	return "sale-" + strconv.FormatInt(ud.Start.Unix(), 10)
}

func (ud upcomingSale) Started() bool {
	return ud.Start.Unix() < time.Now().Unix()
}

func (ud upcomingSale) Ended() bool {
	return ud.End.Unix() < time.Now().Unix()
}

func salesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	var filter = D{
		{"offer_end", M{"$gt": time.Now()}},
	}

	//
	var wg sync.WaitGroup

	var offers []mongo.Sale

	// Get rows
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		offers, err = mongo.GetAllOffers(query.getOffset64(), 100, filter)
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

	var code = helpers.GetProductCC(r)

	for _, offer := range offers {

		var val interface{}

		val, ok := offer.AppPrices[code]
		if !ok {
			val = ""
		}

		response.AddRow([]interface{}{
			offer.AppID,   // 0
			offer.AppName, // 1
			offer.AppIcon, // 2
			helpers.GetAppPath(offer.AppID, offer.AppName), // 3
			val,                     // 4
			offer.OfferPercent,      // 5
			offer.AppRating,         // 6
			offer.OfferEnd.String(), // 7
		})
	}

	response.output(w, r)
}
