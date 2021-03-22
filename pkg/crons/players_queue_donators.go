package crons

import (
	"github.com/gamedb/gamedb/pkg/consumers"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"go.mongodb.org/mongo-driver/bson"
)

type AutoPlayerRefreshes struct {
	BaseTask
}

func (c AutoPlayerRefreshes) ID() string {
	return "update-donator-profiles"
}

func (c AutoPlayerRefreshes) Name() string {
	return "Update donator profiles"
}

func (c AutoPlayerRefreshes) Group() TaskGroup {
	return TaskGroupPlayers
}

func (c AutoPlayerRefreshes) Cron() TaskTime {
	return CronTimeAutoPlayerRefreshes
}

func (c AutoPlayerRefreshes) work() (err error) {

	// Get users
	db, err := mysql.GetMySQLClient()
	if err != nil {
		return err
	}

	var users []mysql.User
	db = db.Select([]string{"id"})
	db = db.Where("level >= ?", mysql.UserLevel2)
	db = db.Find(&users)
	if db.Error != nil {
		return db.Error
	}

	// Update players
	var playerIDs []int64
	for _, user := range users {

		playerID := mysql.GetUserSteamID(user.ID)
		if playerID > 0 {

			playerIDs = append(playerIDs, playerID)

			err = consumers.ProducePlayer(consumers.PlayerMessage{ID: playerID, ForceAchievementsRefresh: true}, "crons-donators")
			err = helpers.IgnoreErrors(err, consumers.ErrInQueue)
			if err != nil {
				return err
			}
		}
	}

	// Update groups
	players, err := mongo.GetPlayersByID(playerIDs, bson.M{"primary_clan_id_string": 1})
	if err != nil {
		return err
	}

	for _, v := range players {
		if v.PrimaryGroupID != "" {
			err = consumers.ProduceGroup(consumers.GroupMessage{ID: v.PrimaryGroupID})
			err = helpers.IgnoreErrors(err, consumers.ErrIsBot, consumers.ErrInQueue)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
