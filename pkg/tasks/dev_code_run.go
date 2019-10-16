package tasks

import (
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	mongodb "go.mongodb.org/mongo-driver/mongo"
)

type DevCodeRun struct {
	BaseTask
}

func (c DevCodeRun) ID() string {
	return "run-dev-code"
}

func (c DevCodeRun) Name() string {
	return "Run dev code"
}

func (c DevCodeRun) Cron() string {
	return ""
}

func (c DevCodeRun) work() {

	codes, err := mongo.GetUniquePlayerCountries()
	if err != nil {
		log.Err(err)
		return
	}

	codes = []string{"GB", ""}
	fields := []rankTask{
		{"Level", "level", "ranks." + mongo.RankKeyLevel + "_"},
		// {"Games", "games_count", "ranks." + mongo.RankKeyLevel + "_"},
		// {"Badges", "badges_count", "ranks." + mongo.RankKeyLevel + "_"},
		// {"Time", "play_time", "ranks." + mongo.RankKeyLevel + "_"},
		// {"Friends", "friends_count", "ranks." + mongo.RankKeyLevel + "_"},
	}

	for _, cc := range append(codes, "ALL") {

		if cc == "" {
			cc = "NONE"
		}

		for _, field := range fields {

			log.Info("Field:" + field.name + " CC:" + cc)

			filter := mongo.M{field.readCol: mongo.M{"$exists": true, "$gt": 0}}
			if cc == "ALL" {
				filter = mongo.M{}
			}

			players, err := mongo.GetPlayers(0, 5, mongo.D{{field.readCol, -1}}, filter, mongo.M{"_id": 1}, nil)
			if err != nil {
				log.Warning(err)
				continue
			}

			var writes []mongodb.WriteModel
			for k, v := range players {

				write := mongodb.NewUpdateOneModel()
				write.SetFilter(mongo.M{"_id": v.ID})
				write.SetUpdate(mongo.M{"$set": mongo.M{field.getWriteCol(cc): k + 1}})
				write.SetUpsert(false)

				writes = append(writes, write)
			}

			err = mongo.BulkUpdatePlayers(writes)
			log.Err(err)
		}
	}
}
