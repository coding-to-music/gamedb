package mongo

import (
	"strconv"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AppDLC struct {
	AppID           int    `bson:"app_id"`
	DLCID           int    `bson:"dlc_id"`
	Icon            string `bson:"icon"`
	Name            string `bson:"name"`
	ReleaseDateUnix int64  `bson:"release_date_unix"`
	ReleaseDateNice string `bson:"release_date_nice"`
}

func (dlc AppDLC) BSON() bson.D {

	return bson.D{
		{"_id", dlc.getKey()},
		{"app_id", dlc.AppID},
		{"dlc_id", dlc.DLCID},
		{"icon", dlc.Icon},
		{"name", dlc.Name},
		{"release_date_unix", dlc.ReleaseDateUnix},
		{"release_date_nice", dlc.ReleaseDateNice},
	}
}

func (dlc AppDLC) getKey() string {
	return strconv.Itoa(dlc.AppID) + "-" + strconv.Itoa(dlc.DLCID)
}

func (dlc AppDLC) GetName() string {
	return helpers.GetAppName(dlc.DLCID, dlc.Name)
}

func (dlc AppDLC) GetIcon() (ret string) {
	return helpers.GetAppIcon(dlc.DLCID, dlc.Icon)
}

func (dlc AppDLC) GetPath() string {
	return helpers.GetAppPath(dlc.DLCID, dlc.Name)
}

func GetDLCForApp(offset int64, limit int64, filter bson.D, sort bson.D) (dlcs []AppDLC, err error) {

	cur, ctx, err := Find(CollectionAppDLC, offset, limit, sort, filter, nil, nil)
	if err != nil {
		return dlcs, err
	}

	defer closeCursor(cur, ctx)

	for cur.Next(ctx) {

		dlc := AppDLC{}
		err := cur.Decode(&dlc)
		if err != nil {
			log.ErrS(err, dlc.getKey())
		} else {
			dlcs = append(dlcs, dlc)
		}
	}

	return dlcs, cur.Err()
}

func ReplaceAppDLCs(DLCs []AppDLC) (err error) {

	if len(DLCs) < 1 {
		return nil
	}

	client, ctx, err := getMongo()
	if err != nil {
		return err
	}

	var writes []mongo.WriteModel
	for _, DLC := range DLCs {

		write := mongo.NewReplaceOneModel()
		write.SetFilter(bson.M{"_id": DLC.getKey()})
		write.SetReplacement(DLC.BSON())
		write.SetUpsert(true)

		writes = append(writes, write)
	}

	collection := client.Database(config.C.MongoDatabase).Collection(CollectionAppDLC.String())
	_, err = collection.BulkWrite(ctx, writes, options.BulkWrite())
	return err
}

func DeleteAppDLC(appID int, DLCs []int) (err error) {

	if len(DLCs) < 1 {
		return nil
	}

	client, ctx, err := getMongo()
	if err != nil {
		return err
	}

	keys := bson.A{}
	for _, DLCID := range DLCs {

		dlc := AppDLC{}
		dlc.DLCID = DLCID
		dlc.AppID = appID

		keys = append(keys, dlc.getKey())
	}

	collection := client.Database(config.C.MongoDatabase).Collection(CollectionAppDLC.String())
	_, err = collection.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": keys}})

	return err
}
