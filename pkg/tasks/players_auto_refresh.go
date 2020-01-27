package tasks

import (
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/sql"
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

func (c AutoPlayerRefreshes) Cron() string {
	return CronTimeAutoPlayerRefreshes
}

func (c AutoPlayerRefreshes) work() (err error) {

	// Get users
	db, err := sql.GetMySQLClient()
	if err != nil {
		return err
	}

	var users []sql.User
	db = db.Select([]string{"steam_id", "steam_id"})
	db = db.Where("patreon_level >= ?", 2)
	db = db.Where("steam_id > ?", 0)
	db = db.Find(&users)
	if db.Error != nil {
		return db.Error
	}

	// Update players
	var playerIDs []int64
	for _, user := range users {

		playerID := user.GetSteamID()

		if playerID > 0 {

			playerIDs = append(playerIDs, playerID)

			err = queue.ProducePlayer(queue.PlayerMessage{ID: playerID})
			log.Err(err)
		}
	}

	// Update groups
	players, err := mongo.GetPlayersByID(playerIDs, bson.M{"primary_clan_id_string": 1})
	for _, v := range players {
		if v.PrimaryGroupID != "" {
			err = queue.ProduceGroup(queue.GroupMessage{ID: v.PrimaryGroupID})
			err = helpers.IgnoreErrors(err, queue.ErrIsBot, memcache.ErrInQueue)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
