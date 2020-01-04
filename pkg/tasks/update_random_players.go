package tasks

import (
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

func (c UpdateRandomPlayers) work() (err error) {

	q, err := queue.Channels[rabbit.Producer][queue.QueuePlayers].Inspect()
	if err != nil {
		return err
	}

	if q.Messages > 0 {
		return nil
	}

	players, err := mongo.GetRandomPlayers(10)
	if err != nil {
		return err
	}

	for _, v := range players {
		err = queue.ProducePlayer(queue.PlayerMessage{ID: v.ID, SkipGroups: true})
		err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
		if err != nil {
			log.Err(err)
		}
	}

	return err
}
