package tasks

import (
	"strconv"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/sql"
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

func (c AutoPlayerRefreshes) work() {

	gorm, err := sql.GetMySQLClient()
	if err != nil {
		log.Err(err)
		return
	}

	var users []sql.User
	gorm = gorm.Select([]string{"steam_id", "steam_id"}).Where("patreon_level >= ?", 2).Where("steam_id > ?", 0).Find(&users)
	if gorm.Error != nil {
		log.Err(gorm.Error)
		return
	}

	var playerIDs []int64

	for _, user := range users {

		playerIDs = append(playerIDs, user.SteamID)

		err = queue.ProduceToSteam(queue.SteamPayload{ProfileIDs: []int64{user.SteamID}}, true)
		log.Err(err)
	}

	var groupIDs []string

	players, err := mongo.GetPlayersByID(playerIDs, mongo.M{"primary_clan_id_string": 1})
	for _, v := range players {
		if v.PrimaryClanIDString != "" {
			groupIDs = append(groupIDs, v.PrimaryClanIDString)
		}
	}

	err = queue.ProduceGroup(groupIDs, false)
	log.Err(err)

	log.Info("Auto updated " + strconv.Itoa(len(users)) + " players")
}
