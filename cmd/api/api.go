package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/didip/tollbooth/v6/limiter"
	"github.com/gamedb/gamedb/cmd/api/generated"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/session"
	"github.com/gamedb/gamedb/pkg/config"
	influxHelpers "github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mysql"
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

type Server struct {
}

func (s Server) returnResponse(w http.ResponseWriter, code int, i interface{}) {

	w.WriteHeader(code)

	err := json.NewEncoder(w).Encode(i)
	if err != nil {
		log.ErrS(err)
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
		key = strings.TrimLeft(r.Header.Get(keyField), "Bearer ")
		if key == "" {
			key = session.Get(r, session.SessionUserAPIKey)
		}
	}

	if key == "" {
		s.returnErrorResponse(w, http.StatusUnauthorized, errors.New("empty api key"))
		return
	}

	if !apiKeyRegexp.MatchString(key) {
		s.returnErrorResponse(w, http.StatusUnauthorized, errors.New("invalid api key"))
		return
	}

	// Rate limit
	if lmt.LimitReached(key) {
		s.returnErrorResponse(w, http.StatusTooManyRequests, err)
		return
	}

	// Check user has access to api
	user, err := mysql.GetUserByAPIKey(key)
	if err == mysql.ErrRecordNotFound {

		s.returnErrorResponse(w, http.StatusUnauthorized, errors.New("invalid api key: "+key))
		return

	} else if err != nil {

		s.returnErrorResponse(w, http.StatusInternalServerError, err)
		return

	} else if user.Level <= mysql.UserLevel3 {

		s.returnErrorResponse(w, http.StatusUnauthorized, errors.New("invalid user level"))
		return
	}

	code, response := callback(w, r)

	go func(r *http.Request, code int, key string, user mysql.User) {

		err = s.saveToInflux(r, code, key, user)
		if err != nil {
			log.ErrS(err)
		}

	}(r, code, key, user)

	switch v := response.(type) {
	case string:
		if code == 200 {
			s.returnResponse(w, code, generated.MessageResponse{Message: v})
		} else {
			s.returnErrorResponse(w, code, errors.New(v))
		}
	case error:
		s.returnErrorResponse(w, code, v)
	default:
		s.returnResponse(w, 200, v)
	}
}

func (s Server) saveToInflux(r *http.Request, code int, key string, user mysql.User) (err error) {

	if config.IsLocal() {
		return nil
	}

	_, err = influxHelpers.InfluxWrite(influxHelpers.InfluxRetentionPolicyAllTime, influx.Point{
		Measurement: string(influxHelpers.InfluxMeasurementAPICalls),
		Tags: map[string]string{
			"path":       r.URL.Path,
			"key":        key,
			"user_email": user.Email,
		},
		Fields: map[string]interface{}{
			strconv.Itoa(code): 1,
		},
		Time:      time.Now(),
		Precision: "ms",
	})

	return err
}
