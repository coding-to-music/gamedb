package main

import (
	"net/http"

	"github.com/gamedb/gamedb/cmd/api/generated"
	"github.com/gamedb/gamedb/pkg/backend"
	generatedBackend "github.com/gamedb/gamedb/pkg/backend/generated"
)

func (s Server) GetGames(w http.ResponseWriter, r *http.Request, params generated.GetGamesParams) {

	s.call(w, r, func(w http.ResponseWriter, r *http.Request) (code int, response interface{}) {

		var limit int64 = 10
		if params.Limit != nil && *params.Limit >= 1 && *params.Limit <= 1000 {
			limit = int64(*params.Limit)
		}

		var offset int64 = 0
		if params.Offset != nil {
			offset = int64(*params.Offset)
		}

		payload := &generatedBackend.ListAppsRequest{
			Pagination: &generatedBackend.PaginationRequest{
				Offset: offset,
				Limit:  limit,
			},
		}

		if params.Ids != nil {
			payload.Ids = *params.Ids
		}

		if params.Tags != nil {
			payload.Tags = *params.Tags
		}

		if params.Genres != nil {
			payload.Genres = *params.Genres
		}

		if params.Categories != nil {
			payload.Categories = *params.Categories
		}

		if params.Developers != nil {
			payload.Developers = *params.Developers
		}

		if params.Publishers != nil {
			payload.Publishers = *params.Publishers
		}

		if params.Platforms != nil {
			payload.Platforms = *params.Platforms
		}

		conn, ctx, err := backend.GetClient()
		if err != nil {
			return 500, err
		}

		resp, err := generatedBackend.NewAppsServiceClient(conn).Apps(ctx, payload)
		if err != nil {
			return 500, err
		}

		result := generated.AppsResponse{}
		result.Pagination.Fill(offset, limit, resp.Pagination.GetTotal())

		for _, app := range resp.Apps {

			result.Apps = append(result.Apps, generated.AppSchema{
				Id:   int(app.GetId()),
				Name: app.GetName(),
			})
		}

		return 200, result
	})
}
