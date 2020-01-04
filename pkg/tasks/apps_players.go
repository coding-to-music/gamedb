package tasks

import (
	"strconv"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/queue"
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

	// Check queue size
	q, err := queue.Channels[rabbit.Producer][queue.QueueAppPlayers].Inspect()
	if err != nil {
		return err
	}

	if q.Messages > 1000 {
		return nil
	}

	// Add apps to queue
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

		err = queue.ProduceAppPlayers(queue.AppPlayerMessage{IDs: chunk})
		log.Err(err)
	}

	return nil
}
