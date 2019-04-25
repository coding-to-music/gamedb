package pages

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gamedb/website/cmd/webserver/session"
	"github.com/gamedb/website/pkg/helpers"
	"github.com/gamedb/website/pkg/log"
	"github.com/gamedb/website/pkg/sql"
	"github.com/go-chi/chi"
)

func ProductKeysRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", productKeysHandler)
	r.Get("/product-keys.json", productKeysAjaxHandler)
	return r
}

func productKeysHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{"type", "key", "value"})
	if ret {
		return
	}

	setCacheHeaders(w, time.Hour*24)

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

	err := returnTemplate(w, r, "product_keys", t)
	log.Err(err, r)
}

type productKeysTemplate struct {
	GlobalTemplate
	Key   string
	Value string
	Type  string
}

func productKeysAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	//
	var code = session.GetCountryCode(r)
	var wg sync.WaitGroup
	var productType = query.getSearchString("type")

	// Get products
	var products []extendedRow
	var recordsFiltered int
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
			log.Err("no product type")
			return
		}

		// Search
		key := query.getSearchString("key")
		if key == "" {
			return
		}
		value := query.getSearchString("value")

		gorm = gorm.Select([]string{"id", "name", "icon", "extended->>'$." + key + "' as value"})

		if value == "" {
			gorm = gorm.Where("extended->>'$." + key + "' != ''")
		} else {
			gorm = gorm.Where("extended->>'$."+key+"' = ?", value)
		}

		// Count
		gorm = gorm.Count(&recordsFiltered)
		log.Err(gorm.Error, r)

		// Order, offset, limit
		gorm = gorm.Limit(100)
		gorm = query.setOrderOffsetGorm(gorm, code, map[string]string{})
		gorm = gorm.Order("change_number_date desc")

		// Get rows
		gorm = gorm.Find(&products)
		log.Err(gorm.Error, r)
	}()

	// Get total
	var count int
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		count, err = sql.CountApps()
		log.Err(err, r)

	}()

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = strconv.Itoa(count)
	response.RecordsFiltered = strconv.Itoa(recordsFiltered)
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

	response.output(w, r)
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
