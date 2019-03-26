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
	ID        int       `bson:"_id"`
	CreatedAt time.Time `bson:"created_at"`
	Apps      []int     ``
	Packages  []int     ``
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
	return bson.M{
		"_id":        change.ID,
		"created_at": change.CreatedAt,
		"apps":       apps,
		"packages":   packages,
	}
}

func (change Change) GetName() (name string) {

	return "Change " + strconv.Itoa(change.ID)
}

func (change Change) GetPath() string {
	return "/changes/" + strconv.Itoa(change.ID)
}

func (change Change) GetTimestamp() int64 {
	return change.CreatedAt.Unix()
}

func (change Change) GetNiceDate() string {
	return change.CreatedAt.Format(helpers.DateYearTime)
}

func (change Change) OutputForJSON() (output []interface{}) {

	return []interface{}{
		change.ID,
		change.CreatedAt.Unix(),
		change.CreatedAt.Format(helpers.DateYearTime),
		change.Apps,
		change.Packages,
		change.GetPath(),
	}
}

func GetChange(id int64) (change Change, err error) {

	var item = helpers.MemcacheChangeRow(id)

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &change, func() (interface{}, error) {

		var change Change

		err = FindDocument(CollectionChanges, "_id", id, &change)

		return change, err
	})

	return change, err
}

func GetChanges(offset int64) (changes []Change, err error) {

	client, ctx, err := getMongo()
	if err != nil {
		return changes, err
	}

	c := client.Database(MongoDatabase, options.Database()).Collection(CollectionChanges.String())

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
