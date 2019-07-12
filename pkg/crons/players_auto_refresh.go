package crons

import (
	"strconv"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/sql"
)

type AutoPlayerRefreshes struct {
}

func (c AutoPlayerRefreshes) ID() CronEnum {
	return CronAutoPlayerRefreshes
}

func (c AutoPlayerRefreshes) Name() string {
	return "Update donator profiles"
}

func (c AutoPlayerRefreshes) Config() sql.ConfigType {
	return sql.ConfAutoProfile
}

func (c AutoPlayerRefreshes) Work() {

	started(c)

	gorm, err := sql.GetMySQLClient()
	if err != nil {
		log.Err(err)
		return
	}

	var users []sql.User
	gorm = gorm.Select([]string{"steam_id"}).Where("patreon_level >= ?", 3).Where("steam_id > ?", 0).Find(&users)
	if gorm.Error != nil {
		log.Err(gorm.Error)
		return
	}

	for _, v := range users {
		err := queue.ProducePlayer(v.SteamID)
		log.Err(err)
	}

	cronLogInfo("Auto updated " + strconv.Itoa(len(users)) + " players")

	finished(c)
}
