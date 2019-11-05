package tasks

import (
	"runtime"
	"strconv"
	"time"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
	. "go.mongodb.org/mongo-driver/bson"
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

	// err2 := mongo.UpdateManyUnset(mongo.CollectionPlayers, M{"ranks": 1})
	// log.Info(err2)

	var ranks = map[int64]M{}
	var fields = []rankTask{
		{"level", mongo.RankKeyLevel},
		{"games_count", mongo.RankKeyGames},
		{"badges_count", mongo.RankKeyBadges},
		{"play_time", mongo.RankKeyPlaytime},
		{"friends_count", mongo.RankKeyFriends},
		{"comments_count", mongo.RankKeyComments},
	}

	// Countries
	countryCodes, err := mongo.GetUniquePlayerCountries()
	if err != nil {
		log.Err(err)
		return
	}
	countryCodes = append(countryCodes, mongo.RankCountryAll)

	for k, cc := range countryCodes {

		if cc == "" {
			cc = mongo.RankCountryNone
		}

		log.Info("Country: " + cc + " (" + strconv.Itoa(k+1) + "/" + strconv.Itoa(len(countryCodes)) + ")")

		for _, field := range fields {

			filter := D{{field.readCol, M{"$exists": true, "$gt": 0}}}
			if cc != mongo.RankCountryAll {
				filter = append(filter, bson.E{Key: "country_code", Value: cc})
			}

			players, err := mongo.GetPlayers(0, 0, D{{field.readCol, -1}}, filter, M{"_id": 1})
			if err != nil {
				log.Err(err)
				continue
			}

			for playerK, v := range players {

				key := strconv.Itoa(int(field.writeCol)) + "_" + cc

				if _, ok := ranks[v.ID]; !ok {
					ranks[v.ID] = M{}
				}

				ranks[v.ID][key] = playerK + 1
			}

			time.Sleep(time.Second * 1 / 2)
		}

		runtime.GC()
	}

	// Rank by State
	for _, cc := range []string{"AU", "CA", "FR", "GB", "NZ", "PH", "SI", "US"} {

		stateCodes, err := mongo.GetUniquePlayerStates(cc)
		if err != nil {
			log.Err(err)
			return
		}

		for k, stateCode := range stateCodes {

			log.Info("State: " + stateCode + " (" + strconv.Itoa(k+1) + "/" + strconv.Itoa(len(stateCodes)) + ")")

			for _, field := range fields {

				filter := D{{"country_code", cc}, {"status_code", stateCode}, {field.readCol, M{"$exists": true, "$gt": 0}}}

				players, err := mongo.GetPlayers(0, 0, D{{field.readCol, -1}}, filter, M{"_id": 1})
				if err != nil {
					log.Err(err)
					continue
				}

				for playerK, v := range players {

					key := strconv.Itoa(int(field.writeCol)) + "_s-" + stateCode

					if _, ok := ranks[v.ID]; !ok {
						ranks[v.ID] = M{}
					}

					ranks[v.ID][key] = playerK + 1
				}

				time.Sleep(time.Second * 1 / 2)
			}

			runtime.GC()
		}
	}

	//
	var writes []mongodb.WriteModel
	for playerID, m := range ranks {

		write := mongodb.NewUpdateOneModel()
		write.SetFilter(M{"_id": playerID})
		write.SetUpdate(M{"$set": M{"ranks": m}})
		write.SetUpsert(false)

		writes = append(writes, write)
	}

	chunks := mongo.ChunkWriteModels(writes, 10000)
	for k, chunk := range chunks {

		log.Info("Saving ranks chunk: " + strconv.Itoa(k+1))

		err = mongo.BulkUpdatePlayers(chunk)
		if val, ok := err.(mongodb.BulkWriteException); ok {
			for _, err2 := range val.WriteErrors {
				log.Err(err2, err2.Request)
			}
		} else {
			log.Err(err)
		}

		time.Sleep(time.Second * 1 / 2)
	}
}

type rankTask struct {
	readCol  string
	writeCol mongo.RankKey
}
