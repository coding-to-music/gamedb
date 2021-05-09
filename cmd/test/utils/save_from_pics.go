package utils

import (
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type saveFromPics struct{}

func (saveFromPics) name() string {
	return "save-from-pics"
}

func (saveFromPics) run() {

	var offset int64 = 0
	var limit int64 = 10_000

	for {

		log.InfoS(offset)

		apps, err := mongo.GetApps(offset, limit, bson.D{{Key: "_id", Value: 1}}, bson.D{{Key: "icon", Value: ""}}, bson.M{"common": 1})
		if err != nil {
			log.Err(err.Error())
			return
		}

		for _, app := range apps {

			icon := app.Common.GetValue("icon")
			if icon != "" {

				_, err = mongo.UpdateOne(mongo.CollectionApps, bson.D{{Key: "_id", Value: app.ID}}, bson.D{{Key: "icon", Value: icon}})
				if err != nil {
					log.Err(err.Error())
				}
			}
		}

		if int64(len(apps)) != limit {
			break
		}

		offset += limit
	}

	log.Info("Done")
}
