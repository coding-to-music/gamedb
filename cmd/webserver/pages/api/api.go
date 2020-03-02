package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/Jleagle/session-go/session"
	"github.com/didip/tollbooth/limiter"
	"github.com/gamedb/gamedb/cmd/webserver/pages/api/generated"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/helpers/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
	influx "github.com/influxdata/influxdb1-client"
)

const (
	keyField = "key"
)

var (
	apiKeyRegexp = regexp.MustCompile("^[A-Z0-9]{20}$")

	// Limiter
	ops = limiter.ExpirableOptions{DefaultExpirationTTL: time.Second}
	lmt = limiter.New(&ops).SetMax(1).SetBurst(10)
)

// Shared between all requests
type Server struct {
}

func (s Server) returnResponse(w http.ResponseWriter, code int, i interface{}) {

	w.WriteHeader(code)

	err := json.NewEncoder(w).Encode(i)
	if err != nil {
		log.Err(err)
	}
}

func (s Server) returnErrorResponse(w http.ResponseWriter, code int, err error) {

	s.returnResponse(w, code, generated.MessageResponse{Message: err.Error()})
}

func (s Server) call(w http.ResponseWriter, r *http.Request, callback func(w http.ResponseWriter, r *http.Request) (code int, response interface{})) {

	var err error

	// Check API key
	key := r.URL.Query().Get(keyField)
	if key == "" {
		key = r.Header.Get(keyField)
		if key == "" {
			key, err = session.Get(r, helpers.SessionUserAPIKey)
			if err != nil {
				s.returnErrorResponse(w, http.StatusUnauthorized, err)
				return
			}
		}
	}

	if key == "" {
		s.returnErrorResponse(w, http.StatusBadRequest, errors.New("no key"))
		return
	}

	if !apiKeyRegexp.MatchString(key) {
		s.returnErrorResponse(w, http.StatusBadRequest, errors.New("invalid key"))
		return
	}

	// Rate limit
	if lmt.LimitReached(key) {
		s.returnErrorResponse(w, http.StatusUnauthorized, err)
		return
	}

	// Check user has access to api
	user, err := sql.GetUserFromKeyCache(key)
	if err == sql.ErrRecordNotFound {

		s.returnErrorResponse(w, http.StatusUnauthorized, errors.New("invalid key: "+key))
		return

	} else if err != nil {

		s.returnErrorResponse(w, http.StatusInternalServerError, err)
		return

	} else if user.PatreonLevel < 3 {

		s.returnErrorResponse(w, http.StatusUnauthorized, errors.New("invalid user level"))
		return

	}

	code, response := callback(w, r)

	go func(r *http.Request, code int, key string, user sql.User) {

		err = s.saveToInflux(r, code, key, user)
		if err != nil {
			log.Err(err)
		}

	}(r, code, key, user)

	s.returnResponse(w, 200, response)
}

func (s Server) saveToInflux(r *http.Request, code int, key string, user sql.User) (err error) {

	if config.IsLocal() {
		return nil
	}

	_, err = influxHelper.InfluxWrite(influxHelper.InfluxRetentionPolicyAllTime, influx.Point{
		Measurement: string(influxHelper.InfluxMeasurementAPICalls),
		Tags: map[string]string{
			"path":       r.URL.Path,
			"key":        key,
			"user_email": user.Email,
		},
		Fields: map[string]interface{}{
			strconv.Itoa(code): 1,
		},
		Time:      time.Now(),
		Precision: "u",
	})

	return err
}
