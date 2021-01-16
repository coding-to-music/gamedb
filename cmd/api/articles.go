package main

import (
	"net/http"

	"github.com/gamedb/gamedb/cmd/api/generated"
	"github.com/gamedb/gamedb/pkg/backend"
	generatedBackend "github.com/gamedb/gamedb/pkg/backend/generated"
)

func (s Server) GetArticles(w http.ResponseWriter, _ *http.Request, params generated.GetArticlesParams) {

	var limit int64 = 10
	if params.Limit != nil && *params.Limit >= 1 && *params.Limit <= 1000 {
		limit = int64(*params.Limit)
	}

	var offset int64 = 0
	if params.Offset != nil {
		offset = int64(*params.Offset)
	}

	payload := &generatedBackend.ListArticlesRequest{
		Pagination: &generatedBackend.PaginationRequest{
			Offset: offset,
			Limit:  limit,
		},
	}

	if params.Ids != nil {
		payload.Ids = *params.Ids
	}

	if params.AppIds != nil {
		payload.AppIds = *params.AppIds
	}

	if params.Feed != nil {
		payload.Feed = *params.Feed
	}

	conn, ctx, err := backend.GetClient()
	if err != nil {
		returnErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	resp, err := generatedBackend.NewArticlesServiceClient(conn).List(ctx, payload)
	if err != nil {
		returnErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	result := generated.ArticlesResponse{}
	result.Pagination.Fill(offset, limit, resp.Pagination.GetTotal())

	for _, article := range resp.Articles {

		result.Articles = append(result.Articles, generated.ArticleSchema{
			AppIcon:   article.GetAppIcon(),
			AppId:     article.GetAppID(),
			Author:    article.GetAuthor(),
			Contents:  article.GetContents(),
			Date:      article.GetDate().GetSeconds(),
			Feed:      article.GetFeedName(),
			FeedLabel: article.GetFeedLabel(),
			FeedType:  article.GetFeedType(),
			Icon:      article.GetArticleIcon(),
			Id:        article.GetId(),
			Title:     article.GetTitle(),
			Url:       article.GetUrl(),
		})
	}

	returnResponse(w, http.StatusOK, result)
}
