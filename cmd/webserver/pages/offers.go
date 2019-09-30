package pages

import (
	"net/http"
	"sync"
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/go-chi/chi"
)

func OffersRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", offersHandler)
	r.Get("/sales.json", offersAjaxHandler)
	return r
}

func offersHandler(w http.ResponseWriter, r *http.Request) {

	t := offersTemplate{}
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

	// Wait
	wg.Wait()

	returnTemplate(w, r, "offers", t)
}

type offersTemplate struct {
	GlobalTemplate
	Tags       []sql.Tag
	Categories []sql.Category
}

func offersAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	var filter = mongo.D{
		{"offer_end", mongo.M{"$gt": time.Now()}},
	}

	//
	var wg sync.WaitGroup

	var offers []mongo.Offer

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
		count, err = mongo.CountDocuments(mongo.CollectionAppOffers, nil, 0)
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
		filtered, err = mongo.CountDocuments(mongo.CollectionAppOffers, filter, 0)
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

		response.AddRow([]interface{}{
			offer.AppID,   // 0
			offer.AppName, // 1
			offer.AppIcon, // 2
			helpers.GetAppPath(offer.AppID, offer.AppName), // 3
			offer.AppPrices[code],                          // 4
			offer.OfferPercent,                             // 5
			offer.AppRating,                                // 6
			offer.OfferEnd.String(),                        // 7
		})
	}

	response.output(w, r)
}
