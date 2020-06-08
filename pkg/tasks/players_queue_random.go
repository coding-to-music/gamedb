package tasks

import (
	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
)

type PlayersQueueRandom struct {
	BaseTask
}

func (c PlayersQueueRandom) ID() string {
	return "update-random-players"
}

func (c PlayersQueueRandom) Name() string {
	return "Update random players"
}

func (c PlayersQueueRandom) Cron() string {
	return CronTimeUpdateRandomPlayers
}

func (c PlayersQueueRandom) work() (err error) {

	// Skip if queues have activity
	queues := map[rabbit.QueueName]int{
		queue.QueueApps:     10,
		queue.QueuePackages: 10,
		queue.QueuePlayers:  10,
	}

	var consumers = 1
	for q, limit := range queues {

		c, err := queue.ProducerChannels[q].Inspect()
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
	players, err := mongo.GetRandomPlayers(5 * consumers)
	if err != nil {
		return err
	}

	for _, v := range players {

		err = queue.ProducePlayer(queue.PlayerMessage{ID: v.ID, SkipPlayerGroups: true})
		err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
		if err != nil {
			return err
		}
	}

	return err
}
