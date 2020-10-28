package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gamedb/gamedb/cmd/api/generated"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/session"
	"github.com/gamedb/gamedb/pkg/config"
	influxHelpers "github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mysql"
	influx "github.com/influxdata/influxdb1-client"
)

const keyField = "key"

var apiKeyRegexp = regexp.MustCompile("^[A-Z0-9]{20}$")

type Server struct {
}

func (s Server) returnResponse(w http.ResponseWriter, code int, i interface{}) {

	w.Header().Set("content-type", "application/json")
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

	go s.saveToInflux(r, code, user.ID)

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

func (s Server) saveToInflux(r *http.Request, code int, userID int) {

	if config.IsLocal() {
		return
	}

	_, err := influxHelpers.InfluxWrite(influxHelpers.InfluxRetentionPolicyAllTime, influx.Point{
		Measurement: string(influxHelpers.InfluxMeasurementAPICalls),
		Tags: map[string]string{
			"path":    r.URL.Path,
			"user_id": strconv.Itoa(userID),
			"code":    strconv.Itoa(code),
		},
		Fields: map[string]interface{}{
			"call": 1,
		},
		Time:      time.Now(),
		Precision: "s",
	})

	if err != nil {
		log.ErrS(err)
	}
}
