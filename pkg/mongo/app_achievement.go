package mongo

import (
	"strconv"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AppAchievement struct {
	AppID       int     `bson:"app_id" json:"app_id"`
	Key         string  `bson:"key" json:"key"`
	Name        string  `bson:"name" json:"name"`
	Description string  `bson:"description" json:"description"`
	Icon        string  `bson:"icon" json:"icon"`
	Completed   float64 `bson:"completed" json:"completed"`
	Hidden      bool    `bson:"hidden" json:"hidden"`   // Just no description
	Active      bool    `bson:"active" json:"active"`   // If it's part of the schema response
	Deleted     bool    `bson:"deleted" json:"deleted"` // Not in global resp anymore
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

func (achievement AppAchievement) GetKey() string {
	return strconv.Itoa(achievement.AppID) + "-" + achievement.Key
}

func (achievement AppAchievement) GetIcon() string {
	return helpers.GetAchievementIcon(achievement.AppID, achievement.Icon)
}

func (achievement AppAchievement) GetCompleted() string {
	return helpers.GetAchievementCompleted(achievement.Completed)
}

func (achievement *AppAchievement) Fill(appID int, response steamapi.SchemaForGameAchievement) {

	achievement.AppID = appID
	achievement.Key = response.Name
	achievement.Name = response.DisplayName
	achievement.Icon = helpers.RegexSha1.FindString(response.Icon)
	achievement.Description = response.Description
	achievement.Hidden = bool(response.Hidden)
}

func GetAppAchievements(offset int64, limit int64, filter bson.D, sort bson.D) (achievements []AppAchievement, err error) {

	cur, ctx, err := Find(CollectionAppAchievements, offset, limit, sort, filter, nil, nil)
	if err != nil {
		return achievements, err
	}

	defer func() {
		err = cur.Close(ctx)
		if err != nil {
			log.ErrS(err)
		}
	}()

	for cur.Next(ctx) {

		var appAcievement AppAchievement
		err := cur.Decode(&appAcievement)
		if err != nil {
			log.ErrS(err, appAcievement.GetKey())
		} else {
			achievements = append(achievements, appAcievement)
		}
	}

	return achievements, cur.Err()
}

func ReplaceAppAchievements(achievements []AppAchievement) (err error) {

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
		write.SetFilter(bson.M{"_id": achievement.GetKey()})
		write.SetReplacement(achievement.BSON())
		write.SetUpsert(true)

		writes = append(writes, write)
	}

	c := client.Database(config.C.MongoDatabase).Collection(CollectionAppAchievements.String())

	_, err = c.BulkWrite(ctx, writes, options.BulkWrite())

	return err
}
