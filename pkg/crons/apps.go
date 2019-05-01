package crons

import (
	"strconv"
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/websockets"
)

func AppPlayers() {

	log.Info("Queueing apps for player checks")

	gorm, err := sql.GetMySQLClient()
	if err != nil {
		log.Critical(err)
		return
	}

	gorm = gorm.Select([]string{"id"})
	gorm = gorm.Order("id ASC")
	gorm = gorm.Model(&[]sql.App{})

	var appIDs []int
	gorm = gorm.Pluck("id", &appIDs)
	if gorm.Error != nil {
		log.Critical(gorm.Error)
	}

	// Chunk appIDs
	var chunks [][]int
	for i := 0; i < len(appIDs); i += 10 {
		end := i + 10

		if end > len(appIDs) {
			end = len(appIDs)
		}

		chunks = append(chunks, appIDs[i:end])
	}

	for _, chunk := range chunks {

		err = queue.ProduceAppPlayers(chunk)
		log.Err(err)
	}

	//
	err = sql.SetConfig(sql.ConfAddedAllAppPlayers, strconv.FormatInt(time.Now().Unix(), 10))
	cronLogErr(err)

	page := websockets.GetPage(websockets.PageAdmin)
	page.Send(websockets.AdminPayload{Message: sql.ConfAddedAllAppPlayers + " complete"})

	cronLogInfo("App players cron complete")
}

func ClearUpcomingCache() {

	var mc = helpers.GetMemcache()
	var err error

	err = mc.Delete(helpers.MemcacheUpcomingAppsCount.Key)
	err = helpers.IgnoreErrors(err, helpers.ErrCacheMiss)
	log.Err(err)

	err = mc.Delete(helpers.MemcacheUpcomingPackagesCount.Key)
	err = helpers.IgnoreErrors(err, helpers.ErrCacheMiss)
	log.Err(err)
}
