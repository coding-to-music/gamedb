package tasks

import (
	"strconv"

	"github.com/gamedb/gamedb/pkg/consumers"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
)

type AppPlayers struct {
	BaseTask
}

func (c AppPlayers) ID() string {
	return "app-players"
}

func (c AppPlayers) Name() string {
	return "Check apps for players"
}

func (c AppPlayers) Cron() string {
	return CronTimeAppPlayers
}

func (c AppPlayers) work() (err error) {

	db, err := sql.GetMySQLClient()
	if err != nil {
		return err
	}

	db = db.Select([]string{"id"})
	db = db.Order("id ASC")
	db = db.Model(&[]sql.App{})

	var appIDs []int
	db = db.Pluck("id", &appIDs)
	if db.Error != nil {
		return db.Error
	}

	log.Info("Found " + strconv.Itoa(len(appIDs)) + " apps")

	chunks := helpers.ChunkInts(appIDs, 10)

	for _, chunk := range chunks {

		err = consumers.ProduceAppPlayers(consumers.AppPlayerMessage{IDs: chunk})
		log.Err(err)
	}

	return nil
}
