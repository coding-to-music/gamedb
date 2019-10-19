package tasks

import (
	"sort"
	"strconv"
	"time"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	mongodb "go.mongodb.org/mongo-driver/mongo"
)

type PlayerRanks struct {
	BaseTask
}

func (c PlayerRanks) ID() string {
	return "update-player-ranks"
}

func (c PlayerRanks) Name() string {
	return "Update player ranks"
}

func (c PlayerRanks) Cron() string {
	return CronTimePlayerRanks
}

func (c PlayerRanks) work() {

	// err2 := mongo.UpdateManyUnset(mongo.CollectionPlayers, mongo.M{"ranks": 1})
	// log.Info(err2)

	codes, err := mongo.GetUniquePlayerCountries()
	if err != nil {
		log.Err(err)
		return
	}
	codes = append(codes, mongo.RankCountryAll)

	sort.SliceStable(codes, func(i, j int) bool {
		return codes[i] < codes[j]
	})

	fields := []rankTask{
		{"level", mongo.RankKeyLevel},
		{"games_count", mongo.RankKeyGames},
		{"badges_count", mongo.RankKeyBadges},
		{"play_time", mongo.RankKeyPlaytime},
		{"friends_count", mongo.RankKeyFriends},
		{"comments_count", mongo.RankKeyComments},
	}

	for _, cc := range codes {

		log.Info("CC: " + cc)

		for _, field := range fields {

			filter := mongo.M{field.readCol: mongo.M{"$exists": true, "$gt": 0}}
			if cc != mongo.RankCountryAll {
				filter["country_code"] = cc
			}

			players, err := mongo.GetPlayers(0, 0, mongo.D{{field.readCol, -1}}, filter, mongo.M{"_id": 1}, nil)
			if err != nil {
				log.Warning(err)
				continue
			}

			time.Sleep(time.Second * 1)

			var writes []mongodb.WriteModel
			for k, v := range players {

				write := mongodb.NewUpdateOneModel()
				write.SetFilter(mongo.M{"_id": v.ID})
				write.SetUpdate(mongo.M{"$set": mongo.M{field.getWriteCol(cc): k + 1}})
				write.SetUpsert(false)

				writes = append(writes, write)
			}

			err = mongo.BulkUpdatePlayers(writes)
			if val, ok := err.(mongodb.BulkWriteException); ok {
				for _, v := range val.WriteErrors {
					log.Err(v)
				}
			} else {
				log.Err(err)
			}

			time.Sleep(time.Second * 1)
		}
	}
}

type rankTask struct {
	readCol  string
	writeCol mongo.RankKey
}

func (rt rankTask) getWriteCol(cc string) string {
	if cc == "" {
		cc = mongo.RankCountryNone
	}
	return "ranks." + strconv.Itoa(int(rt.writeCol)) + "_" + cc
}
