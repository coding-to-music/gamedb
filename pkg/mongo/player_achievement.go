package mongo

import (
	"strconv"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PlayerAchievement struct {
	PlayerID               int64  `bson:"player_id"`
	AppID                  int    `bson:"app_id"`
	AchievementID          string `bson:"achievement_id"`
	AchievementName        string `bson:"achievement_name"`
	AchievementDescription string `bson:"achievement_description"`
	AchievementDate        int64  `bson:"achievement_date"`
}

func (a PlayerAchievement) BSON() bson.D {

	return bson.D{
		{"_id", a.getKey()},
		{"player_id", a.PlayerID},
		{"app_id", a.AppID},
		{"achievement_id", a.AchievementID},
		{"achievement_name", a.AchievementName},
		{"achievement_description", a.AchievementDescription},
		{"achievement_date", a.AchievementDate},
	}
}

func (a PlayerAchievement) getKey() string {
	return strconv.FormatInt(a.PlayerID, 10) + "-" + strconv.Itoa(a.AppID) + "-" + a.AchievementID
}

func FindLatestPlayerAchievement(playerID int64, appID int) (int64, error) {

	var filter = bson.D{{"player_id", playerID}, {"app_id", appID}}

	var playerAchievement PlayerAchievement

	err := FindOne(CollectionPlayerAchievements, filter, bson.D{{"achievement_date", -1}}, bson.M{"achievement_date": 1}, &playerAchievement)
	err = helpers.IgnoreErrors(err, ErrNoDocuments)

	return playerAchievement.AchievementDate, err

}

func GetPlayerAchievements(appID int, offset int64, limit int64, sort bson.D) (achievements []PlayerAchievement, err error) {

	var filter = bson.D{{"app_id", appID}}

	cur, ctx, err := Find(CollectionPlayerAchievements, offset, limit, sort, filter, nil, nil)
	if err != nil {
		return achievements, err
	}

	defer func() {
		err = cur.Close(ctx)
		log.Err(err)
	}()

	for cur.Next(ctx) {

		achievement := PlayerAchievement{}
		err := cur.Decode(&achievement)
		if err != nil {
			log.Err(err, achievement.getKey())
		} else {
			achievements = append(achievements, achievement)
		}
	}

	return achievements, cur.Err()
}

func UpdatePlayerAchievements(achievements []PlayerAchievement) (err error) {

	if len(achievements) < 1 {
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

	collection := client.Database(MongoDatabase).Collection(CollectionPlayerAchievements.String())
	_, err = collection.BulkWrite(ctx, writes, options.BulkWrite())
	return err
}
