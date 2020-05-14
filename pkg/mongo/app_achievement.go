package mongo

import (
	"strconv"
	"strings"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AppAchievement struct {
	AppID       int     `bson:"app_id" json:"-"`
	Key         string  `bson:"key" json:"-"`
	Name        string  `bson:"name" json:"name"` // Only property JSON needs
	Description string  `bson:"description" json:"-"`
	Icon        string  `bson:"icon" json:"icon"`
	Completed   float64 `bson:"completed" json:"-"`
	Hidden      bool    `bson:"hidden" json:"-"`  // Just no description
	Active      bool    `bson:"active" json:"-"`  // If it's part of the schema response
	Deleted     bool    `bson:"deleted" json:"-"` // Not in global resp anymore
}

func (achievement AppAchievement) BSON() bson.D {

	return bson.D{
		{"app_id", achievement.AppID},
		{"key", achievement.Key},
		{"name", achievement.Name},
		{"description", achievement.Description},
		{"icon", achievement.Icon},
		{"completed", achievement.Completed},
		{"hidden", achievement.Hidden},
		{"active", achievement.Active},
		{"deleted", achievement.Deleted},
	}
}

func (achievement AppAchievement) getKey() string {
	return strconv.Itoa(achievement.AppID) + "-" + achievement.Key
}

func (achievement AppAchievement) GetIcon() string {
	return helpers.GetAchievementIcon(achievement.AppID, achievement.Icon)
}

func (achievement *AppAchievement) SetIcon(url string) {

	url = strings.TrimPrefix(url, helpers.AppIconBase+strconv.Itoa(achievement.AppID)+"/")
	url = strings.TrimSuffix(url, ".jpg")
	achievement.Icon = url
}

func GetAppAchievements(offset int64, limit int64, filter bson.D, sort bson.D) (achievements []AppAchievement, err error) {

	cur, ctx, err := Find(CollectionAppAchievements, offset, limit, sort, filter, nil, nil)
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

func SaveAppAchievements(achievements []AppAchievement) (err error) {

	if len(achievements) == 0 {
		return nil
	}

	client, ctx, err := getMongo()
	if err != nil {
		return err
	}

	var writes []mongo.WriteModel
	for _, achievement := range achievements {

		write := mongo.NewReplaceOneModel()
		write.SetFilter(bson.M{"_id": achievement.getKey()})
		write.SetReplacement(achievement.BSON())
		write.SetUpsert(true)

		writes = append(writes, write)
	}

	c := client.Database(MongoDatabase).Collection(CollectionAppAchievements.String())

	_, err = c.BulkWrite(ctx, writes, options.BulkWrite())

	return err
}
