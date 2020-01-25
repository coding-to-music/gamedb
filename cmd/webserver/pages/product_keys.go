package pages

import (
	"net/http"
	"regexp"
	"sync"

	"github.com/gamedb/gamedb/cmd/webserver/helpers/datatable"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/go-chi/chi"
)

func ProductKeysRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", productKeysHandler)
	r.Get("/product-keys.json", productKeysAjaxHandler)
	return r
}

func productKeysHandler(w http.ResponseWriter, r *http.Request) {

	q := r.URL.Query()

	// Template
	t := productKeysTemplate{}
	t.fill(w, r, "Product Keys", "Search extended and common product keys")
	t.Type = q.Get("type")
	t.Key = q.Get("key")
	t.Value = q.Get("value")

	if t.Type != "app" && t.Type != "package" {
		returnErrorTemplate(w, r, errorTemplate{Code: 400, Message: "Invalid Type."})
		return
	}

	returnTemplate(w, r, "product_keys", t)
}

type productKeysTemplate struct {
	GlobalTemplate
	Key   string
	Value string
	Type  string
}

var keyRegex = regexp.MustCompile("[0-9a-z_]+") // To stop SQL injection

func productKeysAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := datatable.NewDataTableQuery(r, false)

	//
	var wg sync.WaitGroup
	var productType = query.GetSearchString("type")

	// Get products
	var products []extendedRow
	var recordsFiltered int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		gorm, err := sql.GetMySQLClient()
		if err != nil {
			log.Err(err, r)
			return
		}

		if productType == "app" {
			gorm = gorm.Table("apps")
		} else if productType == "package" {
			gorm = gorm.Table("packages")
		} else {
			return
		}

		// Search
		key := query.GetSearchString("key")
		if key == "" || !keyRegex.MatchString(key) {
			return
		}
		value := query.GetSearchString("value")

		gorm = gorm.Select([]string{"id", "name", "icon", "extended->>'$." + key + "' as value"})

		if value == "" {
			gorm = gorm.Where("extended->>'$." + key + "' != ''")
		} else {
			gorm = gorm.Where("extended->>'$."+key+"' = ?", value)
		}

		// Count
		gorm = gorm.Count(&recordsFiltered)
		if gorm.Error != nil {
			log.Err(gorm.Error, r)
		}

		// Order, offset, limit
		gorm = gorm.Limit(100)
		gorm = query.SetOrderOffsetGorm(gorm, nil, "")
		gorm = gorm.Order("change_number_date desc")

		// Get rows
		gorm = gorm.Find(&products)
		if gorm.Error != nil {
			log.Err(gorm.Error, r)
		}
	}()

	// Get total
	var count int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		count, err = mongo.CountDocuments(mongo.CollectionApps, nil, 0)
		if err != nil {
			log.Err(err, r)
		}
	}()

	// Wait
	wg.Wait()

	response := datatable.DataTablesResponse{}
	response.Output()
	response.RecordsTotal = int64(count)
	response.RecordsFiltered = recordsFiltered
	response.Draw = query.Draw

	for _, v := range products {
		response.AddRow([]interface{}{
			v.ID,
			v.Name,
			v.GetIcon(),
			v.GetPath(productType),
			v.Value,
		})
	}

	returnJSON(w, r, response)
}

type extendedRow struct {
	ID    int
	Name  string
	Icon  string
	Value string
}

func (e extendedRow) GetIcon() string {
	return helpers.GetAppIcon(e.ID, e.Icon)
}

func (e extendedRow) GetPath(productType string) string {
	if productType == "app" {
		return helpers.GetAppPath(e.ID, e.Name)
	}
	return helpers.GetPackagePath(e.ID, e.Name)
}
