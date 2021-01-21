package mongo

import (
	"strconv"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AppSameOwners struct {
	AppID     int     `bson:"app_id"`
	SameAppID int     `bson:"same_id"`
	Count     int     `bson:"count"`
	Order     float64 `bson:"order"`
}

func (sameOwner AppSameOwners) BSON() bson.D {

	return bson.D{
		{"_id", sameOwner.GetKey()},
		{"app_id", sameOwner.AppID},
		{"same_id", sameOwner.SameAppID},
		{"count", sameOwner.Count},
		{"order", sameOwner.Order},
	}
}

func (sameOwner AppSameOwners) GetKey() string {
	return strconv.Itoa(sameOwner.AppID) + "-" + strconv.Itoa(sameOwner.SameAppID)
}

func ensureAppSameOwnersIndexes() {

	var indexModels = []mongo.IndexModel{
		{Keys: bson.D{{"app_id", 1}, {"order", -1}}},
	}

	//
	client, ctx, err := getMongo()
	if err != nil {
		log.ErrS(err)
		return
	}

	_, err = client.Database(config.C.MongoDatabase).Collection(CollectionAppSameOwners.String()).Indexes().CreateMany(ctx, indexModels)
	if err != nil {
		log.ErrS(err)
	}
}

func GetAppSameOwners(appID int, limit int64) (sameOwners []AppSameOwners, err error) {

	cur, ctx, err := Find(CollectionAppSameOwners, 0, limit, bson.D{{"order", -1}}, bson.D{{"app_id", appID}}, nil, nil)
	if err != nil {
		return sameOwners, err
	}

	defer closeCursor(cur, ctx)

	for cur.Next(ctx) {

		var sameOwner AppSameOwners
		err := cur.Decode(&sameOwner)
		if err != nil {
			log.ErrS(err, sameOwner.GetKey())
		} else {
			sameOwners = append(sameOwners, sameOwner)
		}
	}

	return sameOwners, cur.Err()
}

func ReplaceAppSameOwners(appID int, sameApps []AppSameOwners) (err error) {

	_, err = DeleteMany(CollectionAppSameOwners, bson.D{{"app_id", appID}})
	if err != nil {
		return err
	}

	if len(sameApps) == 0 {
		return nil
	}

	client, ctx, err := getMongo()
	if err != nil {
		return err
	}

	var writes []mongo.WriteModel
	for _, sameApp := range sameApps {

		if sameApp.SameAppID > 0 {

			sameApp.AppID = appID

			write := mongo.NewInsertOneModel()
			write.SetDocument(sameApp.BSON())

			writes = append(writes, write)
		}
	}

	if len(writes) == 0 {
		return nil
	}

	c := client.Database(config.C.MongoDatabase).Collection(CollectionAppSameOwners.String())

	_, err = c.BulkWrite(ctx, writes, options.BulkWrite())

	return err
}
