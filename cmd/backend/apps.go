package main

import (
	"context"

	"github.com/gamedb/gamedb/pkg/backend"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type AppsServer struct {
	backend.AppsServiceServer
}

func (a AppsServer) Apps(ctx context.Context, request *backend.ListAppsRequest) (response *backend.AppsMongoResponse, err error) {

	filter := bson.D{{}}

	if len(request.GetIds()) > 0 {
		filter = append(filter, bson.E{Key: "_id", Value: bson.M{"$in": request.GetIds()}})
	}

	if len(request.GetTags()) > 0 {
		filter = append(filter, bson.E{Key: "tags", Value: bson.M{"$in": request.GetTags()}})
	}

	if len(request.GetGenres()) > 0 {
		filter = append(filter, bson.E{Key: "genres", Value: bson.M{"$in": request.GetGenres()}})
	}

	if len(request.GetCategories()) > 0 {
		filter = append(filter, bson.E{Key: "categories", Value: bson.M{"$in": request.GetCategories()}})
	}

	if len(request.GetDevelopers()) > 0 {
		filter = append(filter, bson.E{Key: "developers", Value: bson.M{"$in": request.GetDevelopers()}})
	}

	if len(request.GetPublishers()) > 0 {
		filter = append(filter, bson.E{Key: "publishers", Value: bson.M{"$in": request.GetPublishers()}})
	}

	if len(request.GetPlatforms()) > 0 {
		filter = append(filter, bson.E{Key: "platforms", Value: bson.M{"$in": request.GetPlatforms()}})
	}

	var projection = bson.M{
		"id":                  1,
		"name":                1,
		"tags":                1,
		"genres":              1,
		"developers":          1,
		"categories":          1,
		"prices":              1,
		"player_peak_alltime": 1,
		"player_peak_week":    1,
		"player_avg_week":     1,
		"release_date_unix":   1,
		"reviews":             1,
		"reviews_score":       1,
	}

	apps, err := mongo.GetApps(request.GetOffset(), request.GetLimit(), bson.D{{"_id", 1}}, filter, projection)
	if err != nil {
		return nil, err
	}

	total, err := mongo.CountDocuments(mongo.CollectionApps, filter, 0)
	if err != nil {
		return nil, err
	}

	response = &backend.AppsMongoResponse{}
	response.Pagination.Total = total

	for _, v := range apps {
		response.Apps = append(response.Apps, &backend.AppMongoResponse{
			Id:   int32(v.GetID()),
			Name: v.GetName(),
		})
	}

	return nil, nil
}

func (a AppsServer) Search(ctx context.Context, r *backend.SearchAppsRequest) (res *backend.AppsElasticResponse, err error) {

	return nil, nil
}
