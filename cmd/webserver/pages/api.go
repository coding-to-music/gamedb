package pages

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/Jleagle/session-go/session"
	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/go-chi/chi"
	"github.com/jinzhu/gorm"
)

var (
	paramAPIKey = apiCallParam{Name: "key", Type: "string"}
	paramID     = apiCallParam{Name: "id", Type: "int"}
	paramOffset = apiCallParam{Name: "offset", Type: "int"}
	paramLimit  = apiCallParam{Name: "limit", Type: "int"}

	endpoints = []apiCall{
		{
			Title: "App",
			Path:  "app",
			Params: []apiCallParam{
				paramAPIKey,
				paramID,
			},
			Handler: apiAppHandler,
		},
		{
			Title: "App - Players",
			Path:  "app-players",
			Params: []apiCallParam{
				paramAPIKey,
				paramID,
				paramOffset,
				paramLimit,
			},
		},
		{
			Title: "App - Price Changes",
			Path:  "app-prices",
			Params: []apiCallParam{
				paramAPIKey,
				paramID,
				paramOffset,
				paramLimit,
			},
		},
		{
			Title: "App - Reviews",
			Path:  "app-reviews",
			Params: []apiCallParam{
				paramAPIKey,
				paramID,
			},
		},
		{
			Title: "Apps",
			Path:  "apps",
			Params: []apiCallParam{
				paramAPIKey,
				paramOffset,
				paramLimit,
			},
			Handler: apiAppsHandler,
		},
		{
			Title: "Apps - New releases",
			Path:  "new-releases",
			Params: []apiCallParam{
				paramAPIKey,
				paramOffset,
				paramLimit,
			},
		},
		{
			Title: "Apps - Trending",
			Path:  "trending-apps",
			Params: []apiCallParam{
				paramAPIKey,
				paramOffset,
				paramLimit,
			},
		},
		{
			Title: "Apps - Keys",
			Path:  "app-keys",
			Params: []apiCallParam{
				paramAPIKey,
				paramOffset,
				paramLimit,
			},
		},
		{
			Title: "Article",
			Path:  "article",
			Params: []apiCallParam{
				paramAPIKey,
				paramID,
			},
		},
		{
			Title: "Articles",
			Path:  "articles",
			Params: []apiCallParam{
				paramAPIKey,
				paramOffset,
				paramLimit,
			},
		},
		{
			Title: "Bundle",
			Path:  "bundle",
			Params: []apiCallParam{
				paramAPIKey,
				paramID,
			},
		},
		{
			Title: "Bundles",
			Path:  "bundles",
			Params: []apiCallParam{
				paramAPIKey,
				paramOffset,
				paramLimit,
			},
		},
		{
			Title: "Change",
			Path:  "change",
			Params: []apiCallParam{
				paramAPIKey,
				paramID,
			},
		},
		{
			Title: "Changes",
			Path:  "changes",
			Params: []apiCallParam{
				paramAPIKey,
				paramOffset,
				paramLimit,
			},
		},
		{
			Title: "Group",
			Path:  "group",
			Params: []apiCallParam{
				paramAPIKey,
				paramID,
			},
		},
		{
			Title: "Groups",
			Path:  "groups",
			Params: []apiCallParam{
				paramAPIKey,
				paramOffset,
				paramLimit,
			},
		},
		{
			Title: "Package",
			Path:  "package",
			Params: []apiCallParam{
				paramAPIKey,
				paramID,
			},
		},
		{
			Title: "Packages",
			Path:  "packages",
			Params: []apiCallParam{
				paramAPIKey,
				paramOffset,
				paramLimit,
			},
		},
		{
			Title: "Player",
			Path:  "player",
			Params: []apiCallParam{
				paramAPIKey,
				paramID,
			},
		},
		{
			Title: "Player - Games",
			Path:  "player-apps",
			Params: []apiCallParam{
				paramAPIKey,
				paramID,
				paramOffset,
				paramLimit,
			},
		},
		{
			Title: "Player - History",
			Path:  "player-history",
			Params: []apiCallParam{
				paramAPIKey,
				paramID,
				paramOffset,
				paramLimit,
			},
		},
		{
			Title: "Players",
			Path:  "players",
			Params: []apiCallParam{
				paramAPIKey,
				paramOffset,
				paramLimit,
			},
		},
		{
			Title: "Stats - Categories",
			Path:  "steam-stats",
			Params: []apiCallParam{
				paramAPIKey,
			},
		},
		{
			Title: "Stats - Genres",
			Path:  "steam-stats",
			Params: []apiCallParam{
				paramAPIKey,
			},
		},
		{
			Title: "Stats - Publishers",
			Path:  "steam-stats",
			Params: []apiCallParam{
				paramAPIKey,
			},
		},
		{
			Title: "Stats - Steam",
			Path:  "steam-stats",
			Params: []apiCallParam{
				paramAPIKey,
			},
		},
		{
			Title: "Stats - Tags",
			Path:  "steam-stats",
			Params: []apiCallParam{
				paramAPIKey,
			},
		},
	}
)

func APIRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", apiHandler)

	for _, v := range endpoints {
		if v.Handler != nil {
			r.Get("/"+v.Path, v.Handler)
		}
	}

	return r
}

func apiHandler(w http.ResponseWriter, r *http.Request) {

	t := apiTemplate{}
	t.fill(w, r, "API", "A list of API endpoints to access Steam data & Game DB data")
	t.Calls = endpoints

	err := returnTemplate(w, r, "api", t)
	log.Err(err, r)
}

type apiTemplate struct {
	GlobalTemplate
	Calls []apiCall
}

type apiCall struct {
	Title   string
	Path    string
	Params  []apiCallParam
	Handler http.HandlerFunc
}

func (c apiCall) Hashtag() string {
	return regexp.MustCompile("[^a-zA-Z0-9]+").ReplaceAllString(c.Title, "")
}

type apiCallParam struct {
	Name    string
	Type    string
	Default string
}

func (p apiCallParam) InputType() string {
	if helpers.SliceHasString([]string{"int", "uint"}, p.Type) {
		return "number"
	}
	return "text"
}

var errNoKey = errors.New("no key")
var errOverID = errors.New("invalid id")
var errOverLimit = errors.New("over rate limit")
var errInvalidKey = errors.New("invalid key")
var errWrongLevelKey = errors.New("wrong level key")
var errInvalidOffset = errors.New("invalid offset")
var errInvalidLimit = errors.New("invalid limit")

var ops = limiter.ExpirableOptions{DefaultExpirationTTL: time.Second}
var lmt = limiter.New(&ops).SetMax(1).SetBurst(2)

func handleAPICall(r *http.Request) (id int64, offset int64, limit int64, err error) {

	q := r.URL.Query()

	// Get key from url/session
	key := q.Get("key")
	if key == "" {
		key, err = session.Get(r, helpers.SessionUserAPIKey)
		if err != nil {
			return id, offset, limit, err
		}
		if key == "" {
			return id, offset, limit, errNoKey
		}
	}

	if len(key) != 20 {
		return id, offset, limit, errInvalidKey
	}

	// Rate limit
	err = tollbooth.LimitByKeys(lmt, []string{key})
	if err != nil {
		// return id, offset, limit, errOverLimit // todo
	}

	// Check user ahs access to api
	level, err := sql.GetUserFromKeyCache(key)
	if err != nil {
		return id, offset, limit, err
	}
	if level.PatreonLevel < 3 {
		return id, offset, limit, errWrongLevelKey
	}

	// Read ID
	val := q.Get("id")
	if val == "" {
		id = 0
	} else {

		id, err = strconv.ParseInt(val, 10, 64)
		if err != nil {
			return id, offset, limit, errOverID
		}

		if id < 1 {
			return id, offset, limit, errOverID
		}
	}

	// Read offset
	val = q.Get("offset")
	if val == "" {
		offset = 0
	} else {

		offset, err = strconv.ParseInt(val, 10, 64)
		if err != nil {
			return id, offset, limit, errInvalidOffset
		}

		if offset < 0 {
			return id, offset, limit, errInvalidOffset
		}
	}

	// Read limit
	val = q.Get("limit")
	if val == "" {
		limit = 10
	} else {

		limit, err = strconv.ParseInt(val, 10, 64)
		if err != nil {
			return id, offset, limit, errInvalidLimit
		}

		if limit < 1 || limit > 1000 {
			return id, offset, limit, errInvalidLimit
		}
	}

	return id, offset, limit, err
}

func handleAPISQLSingle(r *http.Request, db *gorm.DB) (*gorm.DB, error) {

	id, _, _, err := handleAPICall(r)
	if err != nil {
		return db, err
	}

	db = db.Where("id = ?", id)

	return db, db.Error
}

func handleAPISQLMany(r *http.Request, db *gorm.DB) (*gorm.DB, error) {

	_, offset, limit, err := handleAPICall(r)
	if err != nil {
		return db, err
	}

	db = db.Limit(limit)
	db = db.Offset(offset)

	return db, db.Error
}

type apiApp struct {
	ID         int               `json:"id"`
	Name       string            `json:"name"`
	Tags       []int             `json:"tags"`
	Genres     []int             `json:"genres"`
	Developers []int             `json:"developers"`
	Publishers []int             `json:"publishers"`
	Prices     sql.ProductPrices `json:"prices"`
}

func (apiApp *apiApp) fill(sqlApp sql.App) (err error) {

	apiApp.ID = sqlApp.ID
	apiApp.Name = sqlApp.GetName()
	apiApp.Tags, err = sqlApp.GetTagIDs()
	if err != nil {
		return err
	}
	apiApp.Genres, err = sqlApp.GetGenreIDs()
	if err != nil {
		return err
	}
	apiApp.Developers, err = sqlApp.GetDeveloperIDs()
	if err != nil {
		return err
	}
	apiApp.Publishers, err = sqlApp.GetPublisherIDs()
	if err != nil {
		return err
	}
	apiApp.Prices, err = sqlApp.GetPrices()
	if err != nil {
		return err
	}

	return nil
}

func apiAppHandler(w http.ResponseWriter, r *http.Request) {

	db, err := sql.GetMySQLClient()
	if err != nil {
		log.Err(err)
		return
	}

	db = db.Select([]string{"id", "name", "tags", "genres", "developers", "categories", "prices"})
	db = db.Order("id asc")
	db = db.Table("apps")
	db, err = handleAPISQLSingle(r, db)
	if err != nil {
		log.Err(err)
		return
	}

	var app sql.App
	db = db.Find(&app)
	if db.Error != nil {
		log.Err(db.Error)
		return
	}

	apiApp := apiApp{}
	err = apiApp.fill(app)
	if db.Error != nil {
		log.Err(db.Error)
		return
	}

	b, err := json.Marshal(apiApp)
	if err != nil {
		log.Err(err)
		return
	}

	err = returnJSON(w, r, b)
	log.Err(err)
}

func apiAppsHandler(w http.ResponseWriter, r *http.Request) {

	db, err := sql.GetMySQLClient()
	if err != nil {
		log.Err(err)
		return
	}

	db = db.Select([]string{"id", "name", "tags", "genres", "developers", "categories", "prices"})
	db = db.Order("id asc")
	db, err = handleAPISQLMany(r, db)
	if err != nil {
		log.Err(err)
		return
	}

	var sqlApps []sql.App
	db = db.Find(&sqlApps)
	if db.Error != nil {
		log.Err(db.Error)
		return
	}

	//noinspection GoPreferNilSlice
	var apiApps = []apiApp{}

	for _, v := range sqlApps {
		apiApp := apiApp{}
		err = apiApp.fill(v)
		log.Err(err)

		apiApps = append(apiApps, apiApp)
	}

	b, err := json.Marshal(apiApps)
	if err != nil {
		log.Err(err)
		return
	}

	err = returnJSON(w, r, b)
	log.Err(err)
}

func apiPackageHandler(w http.ResponseWriter, r *http.Request) {

}

func apiPackagesHandler(w http.ResponseWriter, r *http.Request) {

}

func apiBundleHandler(w http.ResponseWriter, r *http.Request) {

}

func apiBundlesHandler(w http.ResponseWriter, r *http.Request) {

}

func apiPlayerHandler(w http.ResponseWriter, r *http.Request) {

}

func apiPlayersHandler(w http.ResponseWriter, r *http.Request) {

}

func apiGroupHandler(w http.ResponseWriter, r *http.Request) {

}

func apiGroupsHandler(w http.ResponseWriter, r *http.Request) {

}
