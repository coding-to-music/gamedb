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
	CronAutoPlayerRefreshes CronEnum = "update-auto-player-refreshes"
	CronClearUpcomingCache  CronEnum = "clear-upcoming-apps-cache"
	CronDevelopers          CronEnum = "update-stats-developers"
	CronGenres              CronEnum = "update-stats-genres"
	CronInstagram           CronEnum = "post-to-instagram"
	CronPlayerRanks         CronEnum = "update-player-ranks"
	CronPublishers          CronEnum = "update-stats-publishers"
	CronSteamClientPlayers  CronEnum = "update-steam-client-players"
	CronTags                CronEnum = "update-stats-tags"
	CronWishlists           CronEnum = "update-wishlist"

	CronRegister = map[CronEnum]CronInterface{
		CronAppPlayers:          AppPlayers{},
		CronAutoPlayerRefreshes: AutoPlayerRefreshes{},
		CronClearUpcomingCache:  ClearUpcomingCache{},
		CronDevelopers:          Developers{},
		CronGenres:              Genres{},
		CronInstagram:           Instagram{},
		CronPlayerRanks:         PlayerRanks{},
		CronPublishers:          Publishers{},
		CronSteamClientPlayers:  SteamClientPlayers{},
		CronTags:                Tags{},
		CronWishlists:           Wishlists{},
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
