package handlers

import (
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/gamedb/gamedb/cmd/frontend/helpers/datatable"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson"
)

func ProductKeysRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", productKeysHandler)
	r.Get("/product-keys.json", productKeysAjaxHandler)

	return r
}

func productKeysHandler(w http.ResponseWriter, r *http.Request) {

	t := productKeysTemplate{}
	t.fill(w, r, "product_keys", "PICS Keys", "Search PICS keys")
	t.addAssetChosen()

	var err error
	t.Keys, err = mysql.GetProductKeys()
	if err != nil {
		log.ErrS(err)
	}

	q := r.URL.Query()

	t.Type = q.Get("type")
	t.Comparator = q.Get("comparator")
	t.Key = q.Get("key")
	t.Value = q.Get("value")

	if t.Type != "packages" {
		t.Type = "apps"
	}

	returnTemplate(w, r, t)
}

type productKeysTemplate struct {
	globalTemplate
	Key        string
	Value      string
	Type       string
	Comparator string
	Keys       []mysql.ProductKey
}

var keyRegex = regexp.MustCompile(`(common|config|extended|ufs)\.[0-9a-z\_\@]{2,}`)

func productKeysAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := datatable.NewDataTableQuery(r, false)

	//
	var productType = query.GetSearchString("type")
	var key = query.GetSearchString("key")
	var comparator = query.GetSearchString("comparator")
	var value = query.GetSearchString("value")

	if !keyRegex.MatchString(key) {
		return
	}

	var filter = bson.D{{key, bson.M{"$exists": true}}}

	switch comparator {
	case "equals":
		filter = append(filter, bson.E{Key: key, Value: value})
	case "notequals":
		filter = append(filter, bson.E{Key: key, Value: bson.M{"$ne": value}})
	case "contains":
		if value == "" {
			break
		}
		filter = append(filter, bson.E{Key: key, Value: bson.M{"$regex": regexp.QuoteMeta(value), "$options": "i"}})
	}

	var wg sync.WaitGroup

	var products []productKeyResult
	wg.Add(1)
	go func() {

		defer wg.Done()

		var projection = bson.M{"_id": 1, "name": 1, "icon": 1, key: 1}

		if productType == "packages" {

			packages, err := mongo.GetPackages(query.GetOffset64(), 100, bson.D{{"_id", 1}}, filter, projection)
			if err != nil {
				log.ErrS(err)
				return
			}

			for _, pack := range packages {
				products = append(products, productKeyResult{
					Product: pack,
					Value:   pack.Extended.GetValue(key),
				})
			}

		} else {

			apps, err := mongo.GetApps(query.GetOffset64(), 100, bson.D{{"_id", 1}}, filter, projection)
			if err != nil {
				log.ErrS(err)
				return
			}

			for _, app := range apps {

				product := productKeyResult{
					Product: app,
				}

				key2 := key
				keyParts := strings.Split(key, ".")
				if len(keyParts) > 1 {
					key2 = keyParts[1]
				}

				switch true {
				case strings.HasPrefix(key, "config."):
					product.Value = app.Config.GetValue(key2)
				case strings.HasPrefix(key, "common."):
					product.Value = app.Common.GetValue(key2)
				case strings.HasPrefix(key, "ufs."):
					product.Value = app.UFS.GetValue(key2)
				case strings.HasPrefix(key, "extended."):
					product.Value = app.Extended.GetValue(key2)
				default:
					continue
				}

				products = append(products, product)
			}
		}
	}()

	var filteredCount int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		if productType == "packages" {
			filteredCount, err = mongo.CountDocuments(mongo.CollectionPackages, filter, 0)
		} else {
			filteredCount, err = mongo.CountDocuments(mongo.CollectionApps, filter, 0)
		}
		if err != nil {
			log.ErrS(err)
			return
		}
	}()

	// Get total
	var count int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		if productType == "packages" {
			count, err = mongo.CountDocuments(mongo.CollectionPackages, nil, 0)
		} else {
			count, err = mongo.CountDocuments(mongo.CollectionApps, nil, 0)
		}
		if err != nil {
			log.ErrS(err)
		}
	}()

	// Wait
	wg.Wait()

	var response = datatable.NewDataTablesResponse(r, query, count, filteredCount, nil)
	for _, v := range products {
		response.AddRow([]interface{}{
			v.Product.GetID(),   // 0
			v.Product.GetName(), // 1
			v.Product.GetIcon(), // 2
			v.Product.GetPath(), // 3
			v.Value,             // 4
		})
	}

	returnJSON(w, r, response)
}

type productKeyResult struct {
	Product helpers.ProductInterface
	Value   string
}
