package mongo

import (
	"strconv"
	"time"

	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Change struct {
	CreatedAt time.Time `bson:"created_at"`
	ChangeID  int       `bson:"_id"`
	Apps      []int     ``
	Packages  []int     ``
}

func (change Change) Key() interface{} {
	return change.ChangeID
}

func (change Change) BSON() (ret interface{}) {

	// Apps
	var apps bson.A
	for _, v := range change.Apps {
		apps = append(apps, v)
	}

	// Packages
	var packages bson.A
	for _, v := range change.Packages {
		packages = append(packages, v)
	}

	// BSON
	m := bson.M{
		"_id":        change.ChangeID,
		"created_at": change.CreatedAt,
		"apps":       apps,
		"packages":   packages,
	}

	return m
}

func (change Change) OutputForJSON() (output []interface{}) {

	return []interface{}{
		change.ChangeID,
		change.CreatedAt.Unix(),
		change.CreatedAt.Format(helpers.DateYearTime),
		change.Apps,
		change.Packages,
		change.GetPath(),
	}
}

func (change Change) GetPath() string {
	return "/changes/" + strconv.Itoa(change.ChangeID)
}

func GetChanges(offset int64) (changes []Change, err error) {

	client, ctx, err := GetMongo()
	if err != nil {
		return changes, err
	}

	c := client.Database(MongoDatabase, options.Database()).Collection(CollectionChanges)

	cur, err := c.Find(ctx, bson.M{}, options.Find().SetLimit(100).SetSkip(offset).SetSort(bson.M{"_id": -1}))
	if err != nil {
		return changes, err
	}

	defer func() {
		err = cur.Close(ctx)
		log.Err(err)
	}()

	for cur.Next(ctx) {

		var change Change
		err := cur.Decode(&change)
		log.Err(err)
		changes = append(changes, change)
	}

	return changes, cur.Err()
}
