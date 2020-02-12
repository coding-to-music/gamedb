package mongo

import (
	"strconv"

	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
)

type AppAchievement struct {
	AppID       int     `bson:"app_id"`
	Key         string  `bson:"key"`
	Name        string  `bson:"name"`
	Description string  `bson:"description"`
	Icon        string  `bson:"icon"`
	Completed   float64 `bson:"completed"`
	Active      bool    `bson:"active"`
}

func (achievement AppAchievement) BSON() bson.D {

	return bson.D{
		{"app_id", achievement.AppID},
		{"key", achievement.Key},
		{"name", achievement.Name},
		{"description", achievement.Description},
		{"icon", achievement.Icon},
		{"completed", achievement.Completed},
		{"active", achievement.Active},
	}
}

func (achievement AppAchievement) getKey() string {
	return strconv.Itoa(achievement.AppID) + "-" + achievement.Key
}

func getAppAchievements(appID int, offset int64) (achievements []AppAchievement, err error) {

	var filter = bson.D{{
		"app_id", appID,
	}}

	cur, ctx, err := Find(CollectionPlayerApps, offset, 100, bson.D{{"completed", -1}}, filter, nil, nil)
	if err != nil {
		return achievements, err
	}

	defer func() {
		err = cur.Close(ctx)
		log.Err(err)
	}()

	for cur.Next(ctx) {

		var appAcievement AppAchievement
		err := cur.Decode(&appAcievement)
		if err != nil {
			log.Err(err, appAcievement.getKey())
		} else {
			achievements = append(achievements, appAcievement)
		}
	}

	return achievements, cur.Err()
}
