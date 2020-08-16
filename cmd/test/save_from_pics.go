package main

import (
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

func saveFromPics() error {

	var offset int64 = 0
	var limit int64 = 10_000

	for {

		zap.S().Info(offset)

		apps, err := mongo.GetApps(offset, limit, bson.D{{"_id", 1}}, bson.D{{"icon", ""}}, bson.M{"common": 1})
		if err != nil {
			return err
		}

		for _, app := range apps {

			icon := app.Common.GetValue("icon")
			if icon != "" {

				_, err = mongo.UpdateOne(mongo.CollectionApps, bson.D{{"_id", app.ID}}, bson.D{{"icon", icon}})
				if err != nil {
					zap.S().Error(err)
				}
			}
		}

		if int64(len(apps)) != limit {
			break
		}

		offset += limit
	}

	zap.S().Info("x")

	return nil
}
