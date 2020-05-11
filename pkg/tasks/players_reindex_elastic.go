package tasks

import (
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type PlayersElastic struct {
	BaseTask
}

func (c PlayersElastic) ID() string {
	return "players-reindex-elastic"
}

func (c PlayersElastic) Name() string {
	return "Reindex all players in Elastic"
}

func (c PlayersElastic) Cron() string {
	return ""
}

func (c PlayersElastic) work() (err error) {

	var offset int64 = 0
	var limit int64 = 10_000

	for {

		var projection = bson.M{
		}

		players, err := mongo.GetPlayers(offset, limit, bson.D{{"_id", 1}}, nil, projection)
		if err != nil {
			return err
		}

		for _, app := range players {

			err = queue.ProducePlayerSearch(queue.PlayersSearchMessage{
				ID:   app.ID,
				Name: app.PersonaName,
				Icon: app.Avatar,
			})
			log.Err(err)
		}

		if int64(len(players)) != limit {
			break
		}

		offset += limit
	}

	return nil
}
