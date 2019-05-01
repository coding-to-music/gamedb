package crons

import (
	"strconv"
	"time"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/websockets"
)

func PlayerRanks() {

	cronLogInfo("PlayerRanks updated started")

	cronLogInfo("Level")
	err := mongo.RankPlayers("level", "level_rank")
	log.Warning(err)

	cronLogInfo("Games")
	err = mongo.RankPlayers("games_count", "games_rank")
	log.Warning(err)

	cronLogInfo("Badges")
	err = mongo.RankPlayers("badges_count", "badges_rank")
	log.Warning(err)

	cronLogInfo("Time")
	err = mongo.RankPlayers("play_time", "play_time_rank")
	log.Warning(err)

	cronLogInfo("Friends")
	err = mongo.RankPlayers("friends_count", "friends_rank")
	log.Warning(err)

	//
	err = sql.SetConfig(sql.ConfRanksUpdated, strconv.FormatInt(time.Now().Unix(), 10))
	cronLogErr(err)

	page := websockets.GetPage(websockets.PageAdmin)
	page.Send(websockets.AdminPayload{Message: sql.ConfRanksUpdated + " complete"})

	cronLogInfo("PlayerRanks updated")
}

func AutoPlayerRefreshes() {

	cronLogInfo("Running auto profile updates")

	gorm, err := sql.GetMySQLClient()
	if err != nil {
		log.Err(err)
		return
	}

	gorm = gorm.Select([]string{"player_id"})
	gorm = gorm.Where("patreon_level >= ?", 3)

	var users []sql.User
	gorm = gorm.Find(&users)
	if gorm.Error != nil {
		log.Err(gorm.Error)
		return
	}

	for _, v := range users {
		err := queue.ProducePlayer(v.PlayerID)
		log.Err(err)
	}
}
