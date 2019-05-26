package crons

import (
	"strconv"
	"time"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/websockets"
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

	CronRegister = map[CronEnum]CronInterface{
		CronAppPlayers:          AppPlayers{},
		CronClearUpcomingCache:  ClearUpcomingCache{},
		CronInstagram:           Instagram{},
		CronPlayerRanks:         PlayerRanks{},
		CronAutoPlayerRefreshes: AutoPlayerRefreshes{},
		CronGenres:              Genres{},
		CronPublishers:          Publishers{},
		CronDevelopers:          Developers{},
		CronTags:                Tags{},
		CronSteamClientPlayers:  SteamClientPlayers{},
	}
)

type CronInterface interface {
	ID() CronEnum
	Name() string
	Config() sql.ConfigType
	Work()
}

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

func started(c CronInterface) {

	cronLogInfo("Cron started: " + string(c.Config()))
}

func finished(c CronInterface) {

	// Save config row
	err := sql.SetConfig(c.Config(), strconv.FormatInt(time.Now().Unix(), 10))
	cronLogErr(err)

	// Send websocket
	page := websockets.GetPage(websockets.PageAdmin)
	page.Send(websockets.AdminPayload{Message: string(c.Config()) + " complete"})

	//
	cronLogInfo("Cron complete: " + string(c.Config()))
}
