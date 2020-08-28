package mongo

import (
	"strconv"
	"time"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PlayerAlias struct {
	PlayerID   int64  `bson:"player_id"`
	PlayerName string `bson:"player_name"`
	Time       int64  `bson:"time"`
}

func (a PlayerAlias) BSON() bson.D {

	return bson.D{
		{"_id", a.getKey()},
		{"player_id", a.PlayerID},
		{"player_name", a.PlayerName},
		{"time", a.Time},
	}
}

func (a PlayerAlias) getKey() string {
	return strconv.FormatInt(a.PlayerID, 10) + "-" + strconv.FormatInt(a.Time, 10)
}

func (a PlayerAlias) GetTime() string {
	return time.Unix(a.Time, 0).Format(helpers.DateYearTime)
}

func ReplacePlayerAliases(aliases []PlayerAlias) (err error) {

	if len(aliases) == 0 {
		return nil
	}

	client, ctx, err := getMongo()
	if err != nil {
		return err
	}

	var writes []mongo.WriteModel
	for _, v := range aliases {

		write := mongo.NewReplaceOneModel()
		write.SetFilter(bson.M{"_id": v.getKey()})
		write.SetReplacement(v.BSON())
		write.SetUpsert(true)

		writes = append(writes, write)
	}

	c := client.Database(config.C.MongoDatabase).Collection(CollectionPlayerAliases.String())

	_, err = c.BulkWrite(ctx, writes, options.BulkWrite())

	return err
}

func GetPlayerAliases(playerID int64, limit int64, afterTimestamp int64) (aliases []PlayerAlias, err error) {

	filter := bson.D{
		{"player_id", playerID},
		{"time", bson.M{"$gt": afterTimestamp}},
	}

	cur, ctx, err := Find(CollectionPlayerAliases, 0, limit, bson.D{{"time", -1}}, filter, nil, nil)
	if err != nil {
		return aliases, err
	}

	defer func() {
		err = cur.Close(ctx)
		if err != nil {
			log.ErrS(err)
		}
	}()

	for cur.Next(ctx) {

		alias := PlayerAlias{}
		err := cur.Decode(&alias)
		if err != nil {
			log.ErrS(err, alias.getKey())
		} else {
			aliases = append(aliases, alias)
		}
	}

	return aliases, cur.Err()
}
