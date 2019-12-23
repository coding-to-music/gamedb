package mongo

import (
	"strconv"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PlayerGroup struct {
	PlayerID     int64  `bson:"player_id"`
	GroupID      string `bson:"group_id"`
	GroupName    string `bson:"group_name"`
	GroupIcon    string `bson:"group_icon"`
	GroupMembers int    `bson:"group_members"`
	GroupType    string `bson:"group_type"`
	GroupPrimary bool   `bson:"group_primary"`
	GroupURL     string `bson:"group_url"`
}

func (group PlayerGroup) BSON() bson.D {

	return bson.D{
		{"_id", group.getKey()},
		{"player_id", group.PlayerID},
		{"group_id", group.GroupID},
		{"group_name", group.GroupName},
		{"group_icon", group.GroupIcon},
		{"group_members", group.GroupMembers},
		{"group_type", group.GroupType},
		{"group_primary", group.GroupPrimary},
		{"group_url", group.GroupURL},
	}
}

func (group PlayerGroup) getKey() string {
	return strconv.FormatInt(group.PlayerID, 10) + "-" + group.GroupID
}

func (group PlayerGroup) GetPath() string {
	return helpers.GetGroupPath(group.GroupID, group.GroupName)
}

func (group PlayerGroup) GetType() string {
	return helpers.GetGroupType(group.GroupType)
}

func (group PlayerGroup) IsOfficial() bool {
	return helpers.IsGroupOfficial(group.GroupType)
}

func (group PlayerGroup) GetURL() string {
	return helpers.GetGroupLink(group.GroupType, group.GroupURL)
}

func (group PlayerGroup) GetName() string {
	return helpers.GetGroupName(group.GroupName, group.GroupID)
}

func (group PlayerGroup) GetIcon() string {
	return helpers.AvatarBase + group.GroupIcon
}

func InsertPlayerGroups(groups []PlayerGroup) (err error) {

	if len(groups) < 1 {
		return nil
	}

	client, ctx, err := getMongo()
	if err != nil {
		return err
	}

	var writes []mongo.WriteModel
	for _, group := range groups {

		write := mongo.NewReplaceOneModel()
		write.SetFilter(bson.M{"_id": group.getKey()})
		write.SetReplacement(group.BSON())
		write.SetUpsert(true)

		writes = append(writes, write)
	}

	collection := client.Database(MongoDatabase).Collection(CollectionPlayerGroups.String())
	_, err = collection.BulkWrite(ctx, writes, options.BulkWrite())
	return err
}

func DeletePlayerGroups(playerID int64, groupIDs []string) (err error) {

	if len(groupIDs) < 1 {
		return nil
	}

	client, ctx, err := getMongo()
	if err != nil {
		return err
	}

	keys := bson.A{}
	for _, groupID := range groupIDs {

		player := PlayerGroup{}
		player.PlayerID = playerID
		player.GroupID = groupID

		keys = append(keys, player.getKey())
	}

	collection := client.Database(MongoDatabase).Collection(CollectionPlayerGroups.String())
	_, err = collection.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": keys}})
	return err
}

func GetPlayerGroups(playerID int64, offset int64, limit int64, sort bson.D) (groups []PlayerGroup, err error) {

	var filter = bson.D{{"player_id", playerID}}

	cur, ctx, err := Find(CollectionPlayerGroups, offset, limit, sort, filter, nil, nil)
	if err != nil {
		return groups, err
	}

	defer func(cur *mongo.Cursor) {
		err = cur.Close(ctx)
		log.Err(err)
	}(cur)

	for cur.Next(ctx) {

		var group PlayerGroup
		err := cur.Decode(&group)
		if err != nil {
			log.Err(err, group.getKey(), cur.Current.String())
		} else {
			groups = append(groups, group)
		}
	}

	return groups, cur.Err()
}
