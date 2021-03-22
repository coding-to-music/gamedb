package main

import (
	"net/http"

	"github.com/gamedb/gamedb/cmd/api/generated"
	"github.com/gamedb/gamedb/pkg/backend"
	generatedBackend "github.com/gamedb/gamedb/pkg/backend/generated"
	"github.com/gamedb/gamedb/pkg/log"
)

func (s Server) GetGamesIdSimilar(w http.ResponseWriter, r *http.Request, id int32) {

	conn, ctx, err := backend.GetClient()
	if err != nil {
		log.ErrS(err)
		returnResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	resp, err := generatedBackend.NewAppsServiceClient(conn).Similar(ctx, &generatedBackend.ListSimilarAppsRequest{AppId: id})
	if err != nil {
		log.ErrS(err)
		returnResponse(w, r, http.StatusInternalServerError, err)
		return
	}

	result := generated.SimilarGamesResponse{}

	for _, app := range resp.Apps {

		result.Games = append(result.Games, generated.SimilarGameSchema{
			AppId:  int(app.GetAppId()),
			Count:  int(app.GetCount()),
			Order:  int(app.GetOrder()),
			Owners: int(app.GetOwners()),
		})
	}

	returnResponse(w, r, http.StatusOK, result)
}
