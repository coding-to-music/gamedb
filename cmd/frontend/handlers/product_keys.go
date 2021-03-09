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

	q := r.URL.Query()

	productType := q.Get("type")
	if productType != "packages" {
		productType = "apps"
	}

	keys, err := mysql.GetProductKeys()
	if err != nil {
		log.ErrS(err)
	}

	// Template
	t := productKeysTemplate{}
	t.fill(w, r, "product_keys", "PICS Keys", "Search PICS keys")
	t.Type = productType
	t.Key = q.Get("key")
	t.Value = q.Get("value")
	t.Keys = keys

	returnTemplate(w, r, t)
}

type productKeysTemplate struct {
	globalTemplate
	Key   string
	Value string
	Type  string
	Keys  []mysql.ProductKey
}

var keyRegex = regexp.MustCompile(`(common|config|extended|ufs)\.[0-9a-z_]+`) // To stop injection

func productKeysAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := datatable.NewDataTableQuery(r, false)

	//
	var productType = query.GetSearchString("type")
	var value = query.GetSearchString("value")
	var key = query.GetSearchString("key")

	if key == "" || !keyRegex.MatchString(key) {
		return
	}

	var filter = bson.D{
		{key, bson.M{"$exists": true}},
		{key, value},
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

			for _, v := range packages {
				products = append(products, productKeyResult{
					ID:    v.ID,
					Name:  v.GetName(),
					Icon:  v.GetIcon(),
					Value: v.Extended.GetValue(key),
				})
			}

		} else {

			apps, err := mongo.GetApps(query.GetOffset64(), 100, bson.D{{"_id", 1}}, filter, projection)
			if err != nil {
				log.ErrS(err)
				return
			}

			for _, v := range apps {

				product := productKeyResult{
					ID:   v.ID,
					Name: v.GetName(),
					Icon: v.GetIcon(),
				}

				key2 := key
				keyParts := strings.Split(key, ".")
				if len(keyParts) > 1 {
					key2 = keyParts[1]
				}

				switch true {
				case strings.HasPrefix(key, "config."):
					product.Value = v.Config.GetValue(key2)
				case strings.HasPrefix(key, "common."):
					product.Value = v.Common.GetValue(key2)
				case strings.HasPrefix(key, "ufs."):
					product.Value = v.UFS.GetValue(key2)
				case strings.HasPrefix(key, "extended."):
					product.Value = v.Extended.GetValue(key2)
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
			v.ID,                   // 0
			v.Name,                 // 1
			v.GetIcon(),            // 2
			v.GetPath(productType), // 3
			v.Value,                // 4
		})
	}

	returnJSON(w, r, response)
}

type productKeyResult struct {
	ID    int
	Name  string
	Icon  string
	Value string
}

func (e productKeyResult) GetIcon() string {
	return helpers.GetAppIcon(e.ID, e.Icon)
}

func (e productKeyResult) GetPath(productType string) string {
	if productType == "app" {
		return helpers.GetAppPath(e.ID, e.Name)
	}
	return helpers.GetPackagePath(e.ID, e.Name)
}
