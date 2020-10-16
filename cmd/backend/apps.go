package main

import (
	"context"
	"sync"

	"github.com/gamedb/gamedb/cmd/backend/helpers"
	"github.com/gamedb/gamedb/pkg/backend/generated"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/olivere/elastic/v7"
	"go.mongodb.org/mongo-driver/bson"
)

type AppsServer struct {
}

func (a AppsServer) Apps(ctx context.Context, request *generated.ListAppsRequest) (response *generated.AppsMongoResponse, err error) {

	filter := bson.D{}

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

	apps, err := mongo.GetApps(request.GetPagination().GetOffset(), request.GetPagination().GetLimit(), bson.D{{"_id", 1}}, filter, projection)
	if err != nil {
		return nil, err
	}

	total, err := mongo.CountDocuments(mongo.CollectionApps, nil, 0)
	if err != nil {
		return nil, err
	}

	filtered, err := mongo.CountDocuments(mongo.CollectionApps, filter, 0)
	if err != nil {
		return nil, err
	}

	response = &generated.AppsMongoResponse{}
	response.Pagination = helpers.MakePaginationResponse(request.GetPagination(), total, filtered)

	for _, app := range apps {
		response.Apps = append(response.Apps, &generated.AppMongoResponse{
			Id:         int32(app.GetID()),
			Name:       app.GetName(),
			Tags:       helpers.IntsToInt32s(app.Tags),
			Categories: helpers.IntsToInt32s(app.Categories),
			Developers: helpers.IntsToInt32s(app.Developers),
			Publishers: helpers.IntsToInt32s(app.Publishers),
			Genres:     helpers.IntsToInt32s(app.Genres),
		})
	}

	return response, err
}

func (a AppsServer) Search(ctx context.Context, request *generated.SearchAppsRequest) (response *generated.AppsElasticResponse, err error) {

	var filters []elastic.Query

	if len(request.GetTypes()) > 0 {
		filters = append(filters, elastic.NewTermsQuery("type", helpers.StringsToInterfaces(request.GetTypes())...))
	}

	if len(request.GetTypes()) > 0 {
		filters = append(filters, elastic.NewTermsQuery("tags", helpers.StringsToInterfaces(request.GetTypes())...))
	}

	if len(request.GetTypes()) > 0 {
		filters = append(filters, elastic.NewTermsQuery("genres", helpers.StringsToInterfaces(request.GetTypes())...))
	}

	if len(request.GetTypes()) > 0 {
		filters = append(filters, elastic.NewTermsQuery("developers", helpers.StringsToInterfaces(request.GetTypes())...))
	}

	if len(request.GetTypes()) > 0 {
		filters = append(filters, elastic.NewTermsQuery("publishers", helpers.StringsToInterfaces(request.GetTypes())...))
	}

	if len(request.GetTypes()) > 0 {
		filters = append(filters, elastic.NewTermsQuery("categories", helpers.StringsToInterfaces(request.GetTypes())...))
	}

	if len(request.GetTypes()) > 0 {
		filters = append(filters, elastic.NewTermsQuery("platforms", helpers.StringsToInterfaces(request.GetTypes())...))
	}

	// prices := query.GetSearchSlice("price")
	// if len(prices) == 2 {
	//
	// 	lowCheck, highCheck := false, false
	//
	// 	q := elastic.NewRangeQuery("prices." + string(code) + ".final")
	//
	// 	low, err := strconv.Atoi(strings.Replace(prices[0], ".", "", 1))
	// 	if err != nil {
	// 		log.ErrS(err)
	// 	}
	// 	if err == nil && low > 0 {
	// 		lowCheck = true
	// 		q.From(low)
	// 	}
	//
	// 	high, err := strconv.Atoi(strings.Replace(prices[1], ".", "", 1))
	// 	if err != nil {
	// 		log.ErrS(err)
	// 	}
	// 	if err == nil && high < 100*100 {
	// 		highCheck = true
	// 		q.To(high)
	// 	}
	//
	// 	if lowCheck || highCheck {
	// 		filters = append(filters, q)
	// 	}
	// }
	//
	// scores := query.GetSearchSlice("score")
	// if len(scores) == 2 {
	//
	// 	lowCheck, highCheck := false, false
	//
	// 	q := elastic.NewRangeQuery("score")
	//
	// 	low, err := strconv.Atoi(strings.TrimSuffix(scores[0], ".00"))
	// 	if err != nil {
	// 		log.ErrS(err)
	// 	}
	// 	if err == nil && low > 0 {
	// 		lowCheck = true
	// 		q.From(low)
	// 	}
	//
	// 	high, err := strconv.Atoi(strings.TrimSuffix(scores[1], ".00"))
	// 	if err != nil {
	// 		log.ErrS(err)
	// 	}
	// 	if err == nil && high < 100 {
	// 		highCheck = true
	// 		q.To(high)
	// 	}
	//
	// 	if lowCheck || highCheck {
	// 		filters = append(filters, q)
	// 	}
	// }

	// Get apps
	var wg sync.WaitGroup

	var apps []elasticsearch.App
	var filtered int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		apps, filtered, err = elasticsearch.SearchAppsAdvanced(int(request.GetPagination().GetOffset()), 100, request.GetSearch(), nil, elastic.NewBoolQuery().Filter(filters...))
		if err != nil {
			log.ErrS(err)
		}
	}()

	// Get count
	var total int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		total, err = mongo.CountDocuments(mongo.CollectionApps, nil, 0)
		if err != nil {
			log.ErrS(err)
		}
	}()

	// Wait
	wg.Wait()

	response = &generated.AppsElasticResponse{}
	response.Pagination = helpers.MakePaginationResponse(request.GetPagination(), total, filtered)

	for _, app := range apps {

		response.Apps = append(response.Apps, &generated.AppElasticResponse{
			AchievementsCounts: int64(app.AchievementsCount),
			AchievementsAvg:    float32(app.AchievementsAvg),
			Aliases:            app.Aliases,
			Categories:         helpers.IntsToInt32s(app.Categories),
			Developers:         helpers.IntsToInt32s(app.Developers),
			Followers:          int64(app.FollowersCount),
			Genres:             helpers.IntsToInt32s(app.Genres),
			Icon:               app.Icon,
			Id:                 int32(app.ID),
			Name:               app.Name,
			NameMarked:         app.NameMarked,
			Platforms:          app.Platforms,
			Players:            int64(app.PlayersCount),
			Publishers:         helpers.IntsToInt32s(app.Publishers),
			ReleaseDate:        app.ReleaseDate,
			Score:              float32(app.Score),
			SearchScore:        float32(app.ReviewScore),
			Tags:               helpers.IntsToInt32s(app.Tags),
			Type:               app.Type,
			Trend:              float32(app.Trend),
			WishlistAvg:        float32(app.WishlistAvg),
			WishlistCount:      int32(app.WishlistCount),
			// AchievementIcons:   app.AchievementsIcons,
			// Prices:        app.Prices,
		})
	}

	return response, err
}
