package mongo

import (
	"strconv"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PlayerFriend struct {
	PlayerID     int64     `bson:"player_id"`
	FriendID     int64     `bson:"friend_id"`
	FriendSince  time.Time `bson:"since"`
	Avatar       string    `bson:"avatar"`
	Name         string    `bson:"name"`
	Games        int       `bson:"games"`
	Level        int       `bson:"level"`
	LoggedOff    time.Time `bson:"logged_off"`
	Relationship string    `bson:"relationship"`
}

func (f PlayerFriend) BSON() bson.D {
	return bson.D{
		{"_id", f.getKey()},
		{"player_id", f.PlayerID},
		{"friend_id", f.FriendID},
		{"since", f.FriendSince},
		{"avatar", f.Avatar},
		{"name", f.Name},
		{"games", f.Games},
		{"level", f.Level},
		{"logged_off", f.LoggedOff},
	}
}

func (f PlayerFriend) getKey() (ret interface{}) {

	return strconv.FormatInt(f.PlayerID, 10) + "-" + strconv.FormatInt(f.FriendID, 10)
}

func (f PlayerFriend) Scanned() bool {
	return !f.LoggedOff.IsZero()
}

func (f PlayerFriend) GetPath() string {
	return helpers.GetPlayerPath(f.FriendID, f.Name)
}

func (f PlayerFriend) GetLoggedOff() string {
	if f.Scanned() {
		return f.LoggedOff.Format(helpers.DateYearTime)
	}
	return "-"
}

func (f PlayerFriend) GetFriendSince() string {
	return f.FriendSince.Format(helpers.DateYearTime)
}

func (f PlayerFriend) GetName() string {
	return helpers.GetPlayerName(f.FriendID, f.Name)
}

func (f PlayerFriend) GetLevel() string {
	if f.Scanned() {
		return humanize.Comma(int64(f.Level))
	}
	return "-"
}

func CountFriends(playerID int64) (count int64, err error) {

	return CountDocuments(CollectionPlayerFriends, bson.D{{"player_id", playerID}}, 0)
}

func DeleteFriends(playerID int64, friends []int64) (err error) {

	if len(friends) < 1 {
		return nil
	}

	client, ctx, err := getMongo()
	if err != nil {
		return err
	}

	keys := bson.A{}
	for _, friendID := range friends {

		friend := PlayerFriend{}
		friend.PlayerID = playerID
		friend.FriendID = friendID

		keys = append(keys, friend.getKey())
	}

	collection := client.Database(MongoDatabase).Collection(CollectionPlayerFriends.String())
	_, err = collection.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": keys}})
	return err
}

func UpdateFriends(friends []*PlayerFriend) (err error) {

	if len(friends) < 1 {
		return nil
	}

	client, ctx, err := getMongo()
	if err != nil {
		return err
	}

	var writes []mongo.WriteModel
	for _, friend := range friends {

		write := mongo.NewReplaceOneModel()
		write.SetFilter(bson.M{"_id": friend.getKey()})
		write.SetReplacement(friend.BSON())
		write.SetUpsert(true)

		writes = append(writes, write)
	}

	collection := client.Database(MongoDatabase).Collection(CollectionPlayerFriends.String())
	_, err = collection.BulkWrite(ctx, writes, options.BulkWrite())
	return err
}

func GetFriends(playerID int64, offset int64, limit int64, sort bson.D) (friends []PlayerFriend, err error) {

	var filter = bson.D{{"player_id", playerID}}

	cur, ctx, err := Find(CollectionPlayerFriends, offset, limit, sort, filter, nil, nil)
	if err != nil {
		return friends, err
	}

	defer func(cur *mongo.Cursor) {
		err = cur.Close(ctx)
		log.Err(err)
	}(cur)

	for cur.Next(ctx) {

		var friend PlayerFriend
		err := cur.Decode(&friend)
		if err != nil {
			log.Err(err, friend.getKey())
		}
		friends = append(friends, friend)
	}

	return friends, cur.Err()
}
