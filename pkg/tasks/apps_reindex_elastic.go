package tasks

import (
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type AppsReindexElastic struct {
	BaseTask
}

func (c AppsReindexElastic) ID() string {
	return "apps-reindex-elastic"
}

func (c AppsReindexElastic) Name() string {
	return "Reindex all apps in Elastic"
}

func (c AppsReindexElastic) Cron() string {
	return ""
}

func (c AppsReindexElastic) work() (err error) {

	var offset int64 = 0
	var limit int64 = 10_000

	for {

		var projection = bson.M{
			"_id":              1,
			"name":             1,
			"icon":             1,
			"player_peak_week": 1,
			"group_followers":  1,
			"reviews_score":    1,
			"prices":           1,
		}

		apps, err := mongo.GetApps(offset, limit, bson.D{{"_id", 1}}, nil, projection, nil)
		if err != nil {
			return err
		}

		for _, app := range apps {

			err = queue.ProduceAppSearch(queue.AppsSearchMessage{
				ID:        app.ID,
				Name:      app.Name,
				Icon:      app.Icon,
				Players:   app.PlayerPeakWeek,
				Followers: app.GroupFollowers,
				Score:     app.ReviewsScore,
				Prices:    app.Prices,
			})
			log.Err(err)
		}

		if int64(len(apps)) != limit {
			break
		}

		offset += limit
	}

	return nil
}
