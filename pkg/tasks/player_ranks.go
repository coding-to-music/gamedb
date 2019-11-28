package tasks

import (
	"runtime"
	"strconv"
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
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

func (c PlayerRanks) work() (err error) {

	// err2 := mongo.UpdateManyUnset(mongo.CollectionPlayers, M{"ranks": 1})
	// log.Info(err2)

	var ranks = map[int64]bson.M{}
	var fields = []rankTask{
		{"level", mongo.RankKeyLevel},
		{"games_count", mongo.RankKeyGames},
		{"badges_count", mongo.RankKeyBadges},
		{"play_time", mongo.RankKeyPlaytime},
		{"friends_count", mongo.RankKeyFriends},
		{"comments_count", mongo.RankKeyComments},
	}

	// Continents
	for k, continent := range helpers.Continents {

		log.Info("Continent: " + continent.Value + " (" + strconv.Itoa(k+1) + "/" + strconv.Itoa(len(helpers.Continents)) + ")")

		for _, field := range fields {

			filter := bson.D{
				{field.readCol, bson.M{"$exists": true, "$gt": 0}},
				{"continent_code", continent.Key},
			}

			players, err := mongo.GetPlayers(0, 0, bson.D{{field.readCol, -1}}, filter, bson.M{"_id": 1})
			if err != nil {
				log.Err(err)
				continue
			}

			for position, v := range players {

				key := strconv.Itoa(int(field.writeCol)) + "_" + continent.Key

				if _, ok := ranks[v.ID]; !ok {
					ranks[v.ID] = bson.M{}
				}

				ranks[v.ID][key] = position + 1
			}

			time.Sleep(time.Second * 1 / 2)
		}

		runtime.GC()
	}

	// Countries
	countryCodes, err := mongo.GetUniquePlayerCountries()
	if err != nil {
		return err
	}
	countryCodes = append(countryCodes, mongo.RankCountryAll)

	for k, cc := range countryCodes {

		if cc == "" {
			cc = mongo.RankCountryNone
		}

		log.Info("Country: " + cc + " (" + strconv.Itoa(k+1) + "/" + strconv.Itoa(len(countryCodes)) + ")")

		for _, field := range fields {

			filter := bson.D{{field.readCol, bson.M{"$exists": true, "$gt": 0}}}
			if cc != mongo.RankCountryAll {
				filter = append(filter, bson.E{Key: "country_code", Value: cc})
			}

			players, err := mongo.GetPlayers(0, 0, bson.D{{field.readCol, -1}}, filter, bson.M{"_id": 1})
			if err != nil {
				log.Err(err)
				continue
			}

			for position, v := range players {

				key := strconv.Itoa(int(field.writeCol)) + "_" + cc

				if _, ok := ranks[v.ID]; !ok {
					ranks[v.ID] = bson.M{}
				}

				ranks[v.ID][key] = position + 1
			}

			time.Sleep(time.Second * 1 / 2)
		}

		runtime.GC()
	}

	// Rank by State
	for k, cc := range mongo.CountriesWithStates {

		stateCodes, err := mongo.GetUniquePlayerStates(cc)
		if err != nil {
			log.Err(err)
			continue
		}

		for k2, stateCode := range stateCodes {

			log.Info("Country: " + cc + " (" + strconv.Itoa(k+1) + "/" + strconv.Itoa(len(mongo.CountriesWithStates)) + ") State: " + stateCode.Key + " (" + strconv.Itoa(k2+1) + "/" + strconv.Itoa(len(stateCodes)) + ")")

			for _, field := range fields {

				filter := bson.D{
					{"country_code", cc}, {"status_code", stateCode},
					{field.readCol, bson.M{"$exists": true, "$gt": 0}},
				}

				players, err := mongo.GetPlayers(0, 0, bson.D{{field.readCol, -1}}, filter, bson.M{"_id": 1})
				if err != nil {
					log.Err(err)
					continue
				}

				for position, v := range players {

					key := strconv.Itoa(int(field.writeCol)) + "_s-" + stateCode.Key

					if _, ok := ranks[v.ID]; !ok {
						ranks[v.ID] = bson.M{}
					}

					ranks[v.ID][key] = position + 1
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
		write.SetFilter(bson.M{"_id": playerID})
		write.SetUpdate(bson.M{"$set": bson.M{"ranks": m}})
		write.SetUpsert(false)

		writes = append(writes, write)
	}

	chunks := mongo.ChunkWriteModels(writes, 10000)
	for k, chunk := range chunks {

		log.Info("Saving ranks chunk: " + strconv.Itoa(k+1) + "/" + strconv.Itoa(len(chunks)))

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

	return nil
}

type rankTask struct {
	readCol  string
	writeCol mongo.RankKey
}
