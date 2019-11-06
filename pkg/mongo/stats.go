package mongo

import (
	"time"

	. "go.mongodb.org/mongo-driver/bson"
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

func (t Stat) BSON() D {

	t.UpdatedAt = time.Now()
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}

	return D{
		{"_id", t.ID},
		{"created_at", t.CreatedAt},
		{"updated_at", t.UpdatedAt},
		{"name", t.Name},
		{"apps", t.Apps},
		{"mean_price", t.MeanPrice},
		{"mean_score", t.MeanScore},
		{"mean_players", t.MeanPlayers},
	}
}

func GetStats(c collection, offset int64, limit int64) (stats []Stat) {
	return stats
}

func UpdateStats(c collection, stats []Stat) (err error) {

	if stats == nil || len(stats) == 0 {
		return nil
	}

	client, ctx, err := getMongo()
	if err != nil {
		return err
	}

	var writes []mongo.WriteModel
	for _, stat := range stats {

		write := mongo.NewReplaceOneModel()
		write.SetFilter(M{"_id": stat.ID})
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

	result := collection.FindOne(ctx, M{}, options.FindOne())
	if result.Err() != nil && err != ErrNoDocuments {
		return 0, result.Err()
	}

	err = result.Decode(&row)
	if err != nil {
		return 0, err
	}

	return row.ID, nil
}
