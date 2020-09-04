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
)

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
		apps, filtered, err = elasticsearch.SearchAppsAdvanced(int(request.GetPagination().GetOffset()), request.GetSearch(), nil, filters)
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
