package api

import (
	"errors"
	"net/http"

	"github.com/gamedb/gamedb/cmd/webserver/pages/api/generated"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
)

func (s Server) GetAppsId(w http.ResponseWriter, r *http.Request) {

	s.call(w, r, func(w http.ResponseWriter, r *http.Request) (code int, response interface{}) {

		if id, ok := r.Context().Value("id").(int32); ok {

			app, err := mongo.GetApp(int(id))
			if err == mongo.ErrNoDocuments {

				return 404, errors.New("app not found")

			} else if err != nil {

				log.Err(err)
				return 500, err

			} else {

				ret := generated.AppResponse{}
				ret.Id = app.ID
				ret.Name = app.GetName()

				return 200, ret
			}
		}

		return 400, errors.New("invalid app ID")
	})
}
