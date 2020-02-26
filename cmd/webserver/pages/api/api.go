package api

import (
	"encoding/json"
	"net/http"

	"github.com/gamedb/gamedb/cmd/webserver/pages/api/generated"
	"github.com/gamedb/gamedb/pkg/log"
)

type Server struct {
}

func (s Server) ReturnError(w http.ResponseWriter, code int, message string) {

	w.WriteHeader(code)

	e := generated.ErrorResponse{Code: code, Message: message,}

	err := json.NewEncoder(w).Encode(e)
	log.Err(err)
}

func (s Server) Return200(w http.ResponseWriter, i interface{}) {

	w.WriteHeader(http.StatusOK)

	err := json.NewEncoder(w).Encode(i)
	log.Err(err)
}
