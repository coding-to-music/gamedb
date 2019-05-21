package crons

import (
	"strconv"

	"github.com/gamedb/gamedb/pkg/log"
)

type CronEnum string

var (
	CronAppPlayers          CronEnum = "update-app-players"
	CronClearUpcomingCache  CronEnum = "clear-upcoming-apps-cache"
	CronInstagram           CronEnum = "post-to-instagram"
	CronPlayerRanks         CronEnum = "update-player-ranks"
	CronAutoPlayerRefreshes CronEnum = "update-auto-player-refreshes"
	CronSteamClientPlayers  CronEnum = "update-steam-client-players"
	CronGenres              CronEnum = "update-stats-genres"
	CronPublishers          CronEnum = "update-stats-publishers"
	CronDevelopers          CronEnum = "update-stats-developers"
	CronTags                CronEnum = "update-stats-tags"

	CronRegister = map[CronEnum]func(){
		CronAppPlayers:          AppPlayers,
		CronClearUpcomingCache:  ClearUpcomingCache,
		CronInstagram:           Instagram,
		CronPlayerRanks:         PlayerRanks,
		CronAutoPlayerRefreshes: AutoPlayerRefreshes,
		CronGenres:              Genres,
		CronPublishers:          Publishers,
		CronDevelopers:          Developers,
		CronTags:                Tags,
		CronSteamClientPlayers:  SteamClientPlayers,
	}
)

// type CronInterface interface {
// 	ID() CronEnum
// 	Name() string
// 	Work()
// }

// Logging
func cronLogErr(interfaces ...interface{}) {
	log.Err(append(interfaces, log.LogNameCron, log.LogNameGameDB)...)
}

func cronLogInfo(interfaces ...interface{}) {
	log.Info(append(interfaces, log.LogNameCron, log.LogNameGameDB)...)
}

func statsLogger(tableName string, count int, total int, rowName string) {
	cronLogInfo("Updating " + tableName + " - " + strconv.Itoa(count) + " / " + strconv.Itoa(total) + ": " + rowName)
}

//
