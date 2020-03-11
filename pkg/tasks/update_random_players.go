package tasks

import (
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
)

type UpdateRandomPlayers struct {
	BaseTask
}

func (c UpdateRandomPlayers) ID() string {
	return "update-random-players"
}

func (c UpdateRandomPlayers) Name() string {
	return "Update random players"
}

func (c UpdateRandomPlayers) Cron() string {
	return CronTimeUpdateRandomPlayers
}

const (
	cronInterval = time.Minute
	playerCount  = 5 // Per consumer
)

func (c UpdateRandomPlayers) work() (err error) {

	// Skip if queues have activity
	queues := map[rabbit.QueueName]int{
		queue.QueueApps:     50,
		queue.QueueGroups:   50,
		queue.QueuePackages: 50,
		queue.QueuePlayers:  0,
		queue.QueueDelay:    0,
	}

	var consumers = 1
	for q, limit := range queues {

		c, err := queue.Channels[rabbit.Producer][q].Inspect()
		if err != nil {
			return err
		}

		if c.Messages > limit {
			return nil
		}

		if q == queue.QueuePlayers {
			consumers = c.Consumers
		}
	}

	// Queue players
	players, err := mongo.GetRandomPlayers(playerCount)
	if err != nil {
		return err
	}

	for _, v := range players {
		err = queue.ProducePlayer(queue.PlayerMessage{ID: v.ID, SkipGroups: true})
		err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
		if err != nil {
			log.Err(err)
		}

		time.Sleep(cronInterval / playerCount * time.Duration(consumers))
	}

	return err
}
