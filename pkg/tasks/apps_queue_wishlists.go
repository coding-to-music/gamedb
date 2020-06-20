package tasks

import (
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type AppsWishlistsYoutube struct {
	BaseTask
}

func (c AppsWishlistsYoutube) ID() string {
	return "apps-queue-wishlists"
}

func (c AppsWishlistsYoutube) Name() string {
	return "Update wishlist stats for all apps"
}

func (c AppsWishlistsYoutube) Cron() string {
	return CronTimeAppsWishlists
}

func (c AppsWishlistsYoutube) work() (err error) {

	var offset int64 = 0
	var limit int64 = 10_000

	for {

		apps, err := mongo.GetApps(offset, limit, bson.D{{"_id", 1}}, nil, bson.M{"_id": 1})
		if err != nil {
			return err
		}

		for _, app := range apps {

			err = queue.ProduceAppsWishlists(app.ID)
			if err != nil {
				return err
			}
		}

		if int64(len(apps)) != limit {
			break
		}

		offset += limit
	}

	return nil
}
