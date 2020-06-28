package mongo

import (
	"strconv"

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

func UpdatePlayerAliases(aliases []PlayerAlias) (err error) {

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
		write.SetReplacement(bson.M{"$set": v.BSON()})
		write.SetUpsert(true)

		writes = append(writes, write)
	}

	c := client.Database(MongoDatabase).Collection(CollectionPlayerAliases.String())

	_, err = c.BulkWrite(ctx, writes, options.BulkWrite())

	return err
}
