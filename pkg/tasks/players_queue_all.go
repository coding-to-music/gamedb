package tasks

import (
	"strconv"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
)

type PlayersQueueAll struct {
}

func (c PlayersQueueAll) ID() string {
	return "queue-all-players"
}

func (c PlayersQueueAll) Name() string {
	return "Queue all players"
}

func (c PlayersQueueAll) Cron() string {
	return ""
}

func (c PlayersQueueAll) work() {

	players, err := mongo.GetPlayers(0, 0, mongo.D{{"_id", 1}}, nil, mongo.M{"_id": 1}, nil)
	if err != nil {
		log.Err(err)
		return
	}

	var playerIDs []int64
	for _, player := range players {
		playerIDs = append(playerIDs, player.ID)
	}

	err = queue.ProduceToSteam(queue.SteamPayload{ProfileIDs: playerIDs})
	log.Err(err)

	//
	log.Info(strconv.Itoa(len(players)) + " players added to rabbit")
}
