package pages

import (
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
	}
)

func APIRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", apiHandler)

	for _, v := range endpoints {
		if v.Handler != nil {
			r.Get("/app", v.Handler)
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
	Name string
	Type string
}

func (p apiCallParam) InputType() string {
	if helpers.SliceHasString([]string{"int", "uint"}, p.Type) {
		return "number"
	}
	return "text"
}

var errNoKey = errors.New("no key")
var errOverLimit = errors.New("no key")
var errInvalidKey = errors.New("invalid key")
var errWrongLevelKey = errors.New("wrong level key")
var errInvalidOffset = errors.New("invalid offset")
var errInvalidLimit = errors.New("invalid limit")

var ops = limiter.ExpirableOptions{DefaultExpirationTTL: time.Second}
var lmt = limiter.New(&ops).SetMax(1).SetBurst(2)

func handleAPICall(r *http.Request) (id int64, offset int64, limit int64, err error) {

	q := r.URL.Query()

	key := q.Get("key")
	if key == "" {
		key, err := session.Get(r, helpers.SessionUserAPIKey)
		if err != nil {
			return id, offset, limit, err
		}
		if key == "" {
			return id, offset, limit, errNoKey
		}
	}

	err = tollbooth.LimitByKeys(lmt, []string{key})
	if err != nil {
		return id, offset, limit, errOverLimit
	}

	if len(key) != 20 {
		return id, offset, limit, errInvalidKey
	}

	level, err := sql.GetUserLevelWithKey(key)
	if level < 3 {
		// todo, return missing user error etc
		return id, offset, limit, errWrongLevelKey
	}

	val := q.Get("id")
	id, err = strconv.ParseInt(val, 10, 64)
	if err != nil {
		return id, offset, limit, err
	}

	val = q.Get("offset")
	offset, err = strconv.ParseInt(val, 10, 64)
	if err != nil {
		return id, offset, limit, err
	}

	if offset < 0 {
		return id, offset, limit, errInvalidOffset
	}

	val = q.Get("limit")
	limit, err = strconv.ParseInt(val, 10, 64)
	if err != nil {
		return id, offset, limit, err
	}

	if limit < 0 || limit > 1000 {
		return id, offset, limit, errInvalidLimit
	}

	return id, offset, limit, err
}

func handleAPISQLList(r *http.Request, db *gorm.DB) *gorm.DB {

	// id, ofset, limit, err := handleAPICall(r)

	return db

}

func apiAppsHandler(w http.ResponseWriter, r *http.Request) {

	gorm, err := sql.GetMySQLClient()
	if err != nil {
		log.Err(err)
		return
	}

	gorm = gorm.Select([]string{"id", "name", "screenshots", "reviews_score"})
	gorm = gorm.Order("id asc")
	gorm = handleAPISQLList(r, gorm)

	var apps []sql.App
	gorm = gorm.Find(&apps)
	if gorm.Error != nil {
		log.Err(gorm.Error)
		return
	}

}

func apiPackageHandler(w http.ResponseWriter, r *http.Request) {

}

func apiBundleHandler(w http.ResponseWriter, r *http.Request) {

}

func apiPlayerHandler(w http.ResponseWriter, r *http.Request) {

}

func apiGroupHandler(w http.ResponseWriter, r *http.Request) {

}
