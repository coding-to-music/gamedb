package mongo

import (
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
)

type PlayerAchievement struct {
	PlayerID        int64  `bson:"player_id"`
	AppID           int    `bson:"app_id"`
	AchievementID   string `bson:"achievement_id"`
	AchievementName string `bson:"achievement_name"`
}

func (a PlayerAchievement) BSON() bson.D {

	return bson.D{
		{"_id", a.getKey()},
		{"player_id", a.PlayerID},
		{"app_id", a.AppID},
		{"achievement_id", a.AchievementID},
		{"achievement_name", a.AchievementName},
	}
}

func (a PlayerAchievement) getKey() string {
	return strconv.FormatInt(a.PlayerID, 10) + "-" + strconv.Itoa(a.AppID) + "-" + a.AchievementID
}
