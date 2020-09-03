package main

import (
	"context"

	"github.com/gamedb/gamedb/cmd/backend/helpers"
	"github.com/gamedb/gamedb/pkg/backend/generated"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

func (s StatsServer) Tags(ctx context.Context, request *generated.StatsRequest) (*generated.StatsResponse, error) {
	return s.getList(ctx, request, mongo.StatsTypeTags)
}

func (s StatsServer) Genres(ctx context.Context, request *generated.StatsRequest) (*generated.StatsResponse, error) {
	return s.getList(ctx, request, mongo.StatsTypeGenres)
}

func (s StatsServer) Developers(ctx context.Context, request *generated.StatsRequest) (*generated.StatsResponse, error) {
	return s.getList(ctx, request, mongo.StatsTypeDevelopers)
}

func (s StatsServer) Publishers(ctx context.Context, request *generated.StatsRequest) (*generated.StatsResponse, error) {
	return s.getList(ctx, request, mongo.StatsTypePublishers)
}

func (s StatsServer) Categories(ctx context.Context, request *generated.StatsRequest) (*generated.StatsResponse, error) {
	return s.getList(ctx, request, mongo.StatsTypePublishers)
}

func (s StatsServer) getList(ctx context.Context, request *generated.StatsRequest, typex mongo.StatsType) (response *generated.StatsResponse, err error) {

	offset := request.GetPagination().GetOffset()
	limit := request.GetPagination().GetLimit()

	stats, err := mongo.GetStats(typex, offset, limit)
	if err != nil {
		return nil, err
	}

	total, err := mongo.CountDocuments(mongo.CollectionStats, bson.D{{"type", typex}}, 0)
	if err != nil {
		return nil, err
	}

	response = &generated.StatsResponse{}
	response.Pagination = helpers.MakePagination(request.GetPagination(), total)

	for _, stat := range stats {
		response.Items = append(response.Items, &generated.StatResponse{
			Id:          int32(stat.ID),
			Name:        stat.Name,
			Apps:        int32(stat.Apps),
			MeanPrice:   stat.MeanPrice,
			MeanScore:   stat.MeanScore,
			MeanPlayers: float32(stat.MeanPlayers),
		})
	}

	return response, err
}
