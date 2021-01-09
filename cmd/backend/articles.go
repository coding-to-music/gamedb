package main

import (
	"context"

	"github.com/gamedb/gamedb/cmd/backend/helpers"
	"github.com/gamedb/gamedb/pkg/backend/generated"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/golang/protobuf/ptypes"
	"go.mongodb.org/mongo-driver/bson"
)

type ArticlesServer struct {
	generated.UnimplementedArticlesServiceServer
}

func (as ArticlesServer) List(ctx context.Context, request *generated.ListArticlesRequest) (response *generated.ArticlesResponse, err error) {

	sort := helpers.MakeMongoOrder(request.GetPagination())
	// projection := helpers.MakeMongoProjection(request.GetProjection())

	filter := bson.D{}

	if len(request.GetIds()) > 0 {
		filter = append(filter, bson.E{Key: "_id", Value: bson.M{"$in": request.GetIds()}})
	}

	if len(request.GetAppIds()) > 0 {
		filter = append(filter, bson.E{Key: "app_id", Value: request.GetAppIds()})
	}

	if len(request.GetFeed()) > 0 {
		filter = append(filter, bson.E{Key: "feed_name", Value: request.GetFeed()})
	}

	articles, err := mongo.GetArticles(request.GetPagination().GetOffset(), request.GetPagination().GetLimit(), sort, filter)
	if err != nil {
		return nil, err
	}

	total, err := mongo.CountDocuments(mongo.CollectionGroups, nil, 0)
	if err != nil {
		return nil, err
	}

	filtered, err := mongo.CountDocuments(mongo.CollectionGroups, filter, 0)
	if err != nil {
		return nil, err
	}

	response = &generated.ArticlesResponse{}
	response.Pagination = helpers.MakePaginationResponse(request.GetPagination(), total, filtered)

	for _, group := range articles {
		response.Articles = append(response.Articles, as.makeArticle(group))
	}

	return response, err
}

func (as ArticlesServer) makeArticle(m mongo.Article) (r *generated.ArticleResponse) {

	date, err := ptypes.TimestampProto(m.Date)
	if err != nil {
		log.Err(err.Error())
	}

	return &generated.ArticleResponse{
		Id:          m.ID,
		Title:       m.Title,
		Url:         m.URL,
		IsExternal:  m.IsExternal,
		Author:      m.Author,
		Contents:    m.Contents,
		Date:        date,
		FeedLabel:   m.FeedLabel,
		FeedName:    m.FeedName,
		FeedType:    int32(m.FeedType),
		AppID:       int32(m.AppID),
		AppName:     m.AppName,
		AppIcon:     m.AppIcon,
		ArticleIcon: m.ArticleIcon,
	}
}
