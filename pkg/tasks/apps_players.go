package tasks

import (
	"strconv"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/sql"
)

type AppPlayers struct {
}

func (c AppPlayers) ID() string {
	return "app-players"
}

func (c AppPlayers) Name() string {
	return "Check apps for players"
}

func (c AppPlayers) Cron() string {
	return "@every 5h30m"
}

func (c AppPlayers) work() {

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
}
