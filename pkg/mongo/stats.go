package mongo

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Stat struct {
	ID          int                `bson:"_id"`
	CreatedAt   time.Time          `bson:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"`
	Name        string             `bson:"name"`
	Apps        int                `bson:"apps"`
	MeanPrice   map[string]float32 `bson:"mean_price"`
	MeanScore   float32            `bson:"mean_score"`
	MeanPlayers int                `bson:"mean_players"`
}

func (stat Stat) BSON() bson.D {

	stat.UpdatedAt = time.Now()
	if stat.CreatedAt.IsZero() || stat.CreatedAt.Unix() == 0 {
		stat.CreatedAt = time.Now()
	}

	return bson.D{
		{"_id", stat.ID},
		{"created_at", stat.CreatedAt},
		{"updated_at", stat.UpdatedAt},
		{"name", stat.Name},
		{"apps", stat.Apps},
		{"mean_price", stat.MeanPrice},
		{"mean_score", stat.MeanScore},
		{"mean_players", stat.MeanPlayers},
	}
}

func GetStats(c collection, offset int64, limit int64) (stats []Stat) {
	return stats
}

func ReplaceStats(c collection, stats []Stat) (err error) {

	if len(stats) == 0 {
		return nil
	}

	client, ctx, err := getMongo()
	if err != nil {
		return err
	}

	var writes []mongo.WriteModel
	for _, stat := range stats {

		write := mongo.NewReplaceOneModel()
		write.SetFilter(bson.M{"_id": stat.ID})
		write.SetReplacement(stat.BSON())
		write.SetUpsert(true)

		writes = append(writes, write)
	}

	collection := client.Database(MongoDatabase).Collection(c.String())

	_, err = collection.BulkWrite(ctx, writes, options.BulkWrite())
	return err
}

func FindOrCreate(c collection, name string, id int, row Stat) (foundID int, err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return 0, err
	}

	collection := client.Database(MongoDatabase, options.Database()).Collection(CollectionPlayerApps.String())

	result := collection.FindOne(ctx, bson.M{}, options.FindOne())
	if result.Err() != nil && err != ErrNoDocuments {
		return 0, result.Err()
	}

	err = result.Decode(&row)
	if err != nil {
		return 0, err
	}

	return row.ID, nil
}
