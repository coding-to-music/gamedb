package api

import (
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/session-go/session"
	"github.com/didip/tollbooth/limiter"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/influxdata/influxdb1-client"
	"github.com/jinzhu/gorm"
)

var (
	// Limiters
	ops = limiter.ExpirableOptions{DefaultExpirationTTL: time.Second}
	lmt = limiter.New(&ops).SetMax(1).SetBurst(2)

	// Core params
	ParamAPIKey    = APICallParam{Name: "key", Type: "string"}
	ParamPage      = APICallParam{Name: "page", Type: "int"}
	ParamLimit     = APICallParam{Name: "limit", Type: "int"}
	ParamSortField = APICallParam{Name: "sort_field", Type: "string"}
	ParamSortOrder = APICallParam{Name: "sort_order", Type: "string"}
)

type APIRequest struct {
	request   *http.Request
	userEmail string
}

func NewAPICall(r *http.Request) (api APIRequest, err error) {

	call := APIRequest{request: r}

	key, err := call.geKey()
	if err != nil {
		return call, err
	}

	// Rate limit
	if lmt.LimitReached(key) {
		return call, errors.New("rate limit reached")
	}

	// Check user has access to api
	user, err := sql.GetUserFromKeyCache(key)
	if err == sql.ErrRecordNotFound {
		return call, errors.New("invalid key: " + key)
	}
	if err != nil {
		return call, err
	}
	if user.PatreonLevel < 3 {
		return call, errors.New("invalid user level")
	}

	call.userEmail = user.Email

	return call, nil
}

func (r APIRequest) geKey() (key string, err error) {

	key, err = r.getQueryString(ParamAPIKey.Name)
	if err != nil {
		key = r.request.Header.Get(ParamAPIKey.Name)
		if key == "" {
			key, err = session.Get(r.request, helpers.SessionUserAPIKey)
			if err != nil {
				return key, err
			}
		}
	}

	if key == "" {
		return key, errors.New("no key")
	}

	if len(key) != 20 {
		return key, errors.New("invalid key")
	}

	return key, err
}

func (r APIRequest) SaveToInflux(success bool, callError error) (err error) {

	if config.IsLocal() {
		return
	}

	// Fields
	fields := map[string]interface{}{}
	if success {
		fields["success"] = 1
	} else {
		fields["error"] = 1
	}

	// Tags
	key, _ := r.geKey()

	tags := map[string]string{
		"path":       r.request.URL.Path,
		"key":        key,
		"user_email": r.userEmail,
	}

	if callError != nil {
		tags["error"] = callError.Error()
	}

	// Save to Influx
	_, err = helpers.InfluxWrite(helpers.InfluxRetentionPolicyAllTime, client.Point{
		Measurement: string(helpers.InfluxMeasurementAPICalls),
		Tags:        tags,
		Fields:      fields,
		Time:        time.Now(),
		Precision:   "u",
	})

	return err
}

var errParamNotSet = errors.New("param not set")

func (r APIRequest) getQueryString(key string) (val string, err error) {
	val = r.request.URL.Query().Get(key)
	if val == "" {
		return val, errParamNotSet
	}
	return val, nil
}

func (r APIRequest) getQueryInt(key string) (val int64, err error) {
	v := r.request.URL.Query().Get(key)
	if v == "" {
		return 0, errParamNotSet
	} else {
		return strconv.ParseInt(v, 10, 64)
	}
}

func (r APIRequest) setSQLLimitOffset(db *gorm.DB) (*gorm.DB, error) {

	var err error

	// Limit
	limit, err := r.getQueryInt(ParamLimit.Name)
	if err == errParamNotSet {
		limit = 10
	} else if err != nil {
		return db, err
	}
	if limit <= 0 || limit > 1000 {
		return db, errors.New("invalid limit")
	}

	db = db.Limit(limit)

	// Page
	page, err := r.getQueryInt(ParamPage.Name)
	if err == errParamNotSet {
		page = 1
	} else if err != nil {
		return db, err
	}
	if page < 1 {
		return db, errors.New("invalid page")
	}

	db = db.Offset((page - 1) * limit)

	return db, db.Error
}

func (r APIRequest) setSQLOrder(db *gorm.DB, allowed func(in string) (out string)) (*gorm.DB, error) {

	field, err := r.getQueryString(ParamSortField.Name)
	if err != nil {
		field = "id"
	}
	fieldReal := allowed(field)
	if fieldReal == "" {
		return db, errors.New("invalid sort field")
	}

	v, err := r.getQueryString(ParamSortOrder.Name)
	if err != nil {
		v = "asc"
	}

	switch v {
	case "asc":
		db = db.Order(fieldReal + " ASC")
	case "desc":
		db = db.Order(fieldReal + " DESC")
	default:
		return db, errors.New("invalid sort order")
	}

	return db, db.Error
}

func (r APIRequest) setSQLSelect(db *gorm.DB, allowed []string) (*gorm.DB, error) {

	var use []string
	field, err := r.getQueryString("filter")
	if err != nil {
		use = allowed
	} else {
		for _, v := range strings.Split(field, ",") {
			v = strings.TrimSpace(v)
			for _, vv := range allowed {
				if v == vv {
					use = append(use, vv)
					break
				}
			}
		}

		if len(use) == 0 {
			return db, errors.New("invalid filter")
		}
	}

	db = db.Select(use)

	return db, db.Error
}

//
type APICallParam struct {
	Name    string
	Type    string
	Default string
	Comment string
}

func (p APICallParam) InputType() string {
	if helpers.SliceHasString([]string{"int", "uint"}, p.Type) {
		return "number"
	}
	return "text"
}

//
type APICall struct {
	Title   string
	Version int
	Path    string
	Params  []APICallParam
	Handler http.HandlerFunc
}

func (c APICall) Hashtag() string {
	return regexp.MustCompile("[^a-zA-Z0-9]+").ReplaceAllString(c.Title, "")
}

func (c APICall) GetPath() string {
	return "/" + c.VersionString() + "/" + c.Path
}

func (c APICall) VersionString() string {
	if c.Version == 0 {
		c.Version = 1
	}
	return "v" + strconv.Itoa(c.Version)
}
