package crons

import (
	"strconv"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/sql"
)

type AppPlayers struct {
}

func (c AppPlayers) ID() CronEnum {
	return CronAppPlayers
}

func (c AppPlayers) Name() string {
	return "Check apps for players"
}

func (c AppPlayers) Config() sql.ConfigType {
	return sql.ConfAddedAllAppPlayers
}

func (c AppPlayers) Work() {

	started(c)

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

	log.Info("Found " + strconv.Itoa(len(appIDs)) + " apps")

	// Chunk appIDs
	var chunks [][]int
	for i := 0; i < len(appIDs); i += 10 {
		end := i + 10

		if end > len(appIDs) {
			end = len(appIDs)
		}

		chunks = append(chunks, appIDs[i:end])
	}

	log.Info("Chunking")

	for _, chunk := range chunks {

		err = queue.ProduceAppPlayers(chunk)
		log.Err(err)
	}

	finished(c)
}

type ClearUpcomingCache struct {
}

func (c ClearUpcomingCache) ID() CronEnum {
	return CronClearUpcomingCache
}

func (c ClearUpcomingCache) Name() string {
	return "Clear upcoming apps cache"
}

func (c ClearUpcomingCache) Config() sql.ConfigType {
	return sql.ConfClearUpcomingCache
}

func (c ClearUpcomingCache) Work() {

	started(c)

	var err error

	err = helpers.ClearMemcache(helpers.MemcacheUpcomingAppsCount)
	cronLogErr(err)

	err = helpers.ClearMemcache(helpers.MemcacheUpcomingPackagesCount)
	cronLogErr(err)

	finished(c)
}
