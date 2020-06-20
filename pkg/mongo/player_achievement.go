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
	PlayerID               int64   `bson:"player_id"`
	PlayerName             string  `bson:"player_name"`
	PlayerIcon             string  `bson:"player_icon"`
	AppID                  int     `bson:"app_id"`
	AppName                string  `bson:"app_name"`
	AppIcon                string  `bson:"app_icon"`
	AchievementID          string  `bson:"achievement_id"`
	AchievementName        string  `bson:"achievement_name"`
	AchievementIcon        string  `bson:"achievement_icon"`
	AchievementDescription string  `bson:"achievement_description"`
	AchievementDate        int64   `bson:"achievement_date"`
	AchievementComplete    float64 `bson:"achievement_complete"`
}

func (a PlayerAchievement) BSON() bson.D {

	return bson.D{
		{"_id", a.getKey()},
		{"player_id", a.PlayerID},
		{"player_name", a.PlayerName},
		{"player_icon", a.PlayerIcon},
		{"app_id", a.AppID},
		{"app_name", a.AppName},
		{"app_icon", a.AppIcon},
		{"achievement_id", a.AchievementID},
		{"achievement_name", a.AchievementName},
		{"achievement_icon", a.AchievementIcon},
		{"achievement_description", a.AchievementDescription},
		{"achievement_date", a.AchievementDate},
		{"achievement_complete", a.AchievementComplete},
	}
}

func (a PlayerAchievement) getKey() string {
	return strconv.FormatInt(a.PlayerID, 10) + "-" + strconv.Itoa(a.AppID) + "-" + a.AchievementID
}

func (a PlayerAchievement) GetAchievementIcon() string {
	return helpers.GetAchievementIcon(a.AppID, a.AchievementIcon)
}

func (a PlayerAchievement) GetComplete() string {
	return helpers.GetAchievementCompleted(a.AchievementComplete)
}

func FindLatestPlayerAchievement(playerID int64, appID int) (int64, error) {

	var filter = bson.D{{"player_id", playerID}, {"app_id", appID}}

	var playerAchievement PlayerAchievement

	err := FindOne(CollectionPlayerAchievements, filter, bson.D{{"achievement_date", -1}}, bson.M{"achievement_date": 1}, &playerAchievement)
	err = helpers.IgnoreErrors(err, ErrNoDocuments)

	return playerAchievement.AchievementDate, err

}

func GetPlayerAchievements(playerID int64, offset int64) (achievements []PlayerAchievement, err error) {

	var filter = bson.D{{"player_id", playerID}}
	var sort = bson.D{{"achievement_date", -1}}

	return getPlayerAchievements(offset, 100, filter, sort)
}

func GetPlayerAchievementsForApp(playerID int64, appID int, achievementKeys bson.A, limit int64) (achievements []PlayerAchievement, err error) {

	if len(achievementKeys) < 1 {
		return achievements, err
	}

	var filter = bson.D{
		{"player_id", playerID},
		{"app_id", appID},
		{"achievement_id", bson.M{"$in": achievementKeys}},
	}

	return getPlayerAchievements(0, limit, filter, nil)
}

func GetPlayersWithAchievement(appID int, achievementID string, offset int64) (achievements []PlayerAchievement, err error) {

	var filter = bson.D{
		{"app_id", appID},
		{"achievement_id", achievementID},
	}

	return getPlayerAchievements(offset, 100, filter, bson.D{{"achievement_date", -1}})
}

func getPlayerAchievements(offset int64, limit int64, filter bson.D, sort bson.D) (achievements []PlayerAchievement, err error) {

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
