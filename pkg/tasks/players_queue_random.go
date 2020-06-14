package tasks

import (
	"time"

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

const toQueue = 10
const cronTime = time.Minute

func (c PlayersQueueRandom) work() (err error) {

	// Skip if queues have activity
	limits := map[rabbit.QueueName]int{
		queue.QueueApps:     50,
		queue.QueuePackages: 50,
		queue.QueuePlayers:  0,
	}

	queues, err := helpers.RabbitClient.GetQueues()
	if err != nil {
		return err
	}

	var consumers int
	for _, q := range queues {
		if val, ok := limits[rabbit.QueueName(q.Name)]; ok && q.Messages > val {
			return nil
		}
		if q.Name == string(queue.QueuePlayers) {
			consumers = q.Consumers
		}
	}

	// Queue players
	players, err := mongo.GetRandomPlayers(toQueue * consumers)
	if err != nil {
		return err
	}

	for _, v := range players {

		err = queue.ProducePlayer(queue.PlayerMessage{ID: v.ID, SkipPlayerGroups: true})
		err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
		if err != nil {
			return err
		}

		time.Sleep(cronTime / time.Duration(toQueue*consumers))
	}

	return err
}
