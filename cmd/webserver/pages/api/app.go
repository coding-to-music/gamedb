package api

import (
	"net/http"

	"github.com/gamedb/gamedb/cmd/webserver/pages/api/generated"
	"github.com/gamedb/gamedb/pkg/mongo"
)

func (s Server) GetAppsId(w http.ResponseWriter, r *http.Request) {

	if id, ok := r.Context().Value("id").(int); ok {

		app, err := mongo.GetApp(id)
		if err == mongo.ErrNoDocuments {

			s.ReturnError(w, 404, "App not found")

		} else if err != nil {

			s.ReturnError(w, 500, err.Error())

		} else {

			ret := generated.AppResponse{}
			ret.Id = app.ID
			ret.Name = app.GetName()

			s.Return200(w, ret)
		}
	}
}
