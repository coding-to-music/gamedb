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
	PlayerID      int64  `bson:"player_id"`
	PlayerName    string `bson:"player_name"`
	PlayerAvatar  string `bson:"player_avatar"`
	PlayerCountry string `bson:"player_country"`
	PlayerLevel   int    `bson:"player_level"`
	PlayerGames   int    `bson:"player_games"`
	GroupID       string `bson:"group_id"`
	GroupName     string `bson:"group_name"`
	GroupIcon     string `bson:"group_icon"`
	GroupMembers  int    `bson:"group_members"`
	GroupType     string `bson:"group_type"`
	GroupURL      string `bson:"group_url"`
}

func (group PlayerGroup) BSON() bson.D {

	return bson.D{
		{"_id", group.getKey()},
		{"player_id", group.PlayerID},
		{"player_name", group.PlayerName},
		{"player_avatar", group.PlayerAvatar},
		{"player_country", group.PlayerCountry},
		{"player_level", group.PlayerLevel},
		{"player_games", group.PlayerGames},
		{"group_id", group.GroupID},
		{"group_name", group.GroupName},
		{"group_icon", group.GroupIcon},
		{"group_members", group.GroupMembers},
		{"group_type", group.GroupType},
		{"group_url", group.GroupURL},
	}
}

func (group PlayerGroup) getKey() string {
	return strconv.FormatInt(group.PlayerID, 10) + "-" + group.GroupID
}

func (group PlayerGroup) GetGroupPath() string {
	return helpers.GetGroupPath(group.GroupID, group.GetName())
}

func (group PlayerGroup) GetPlayerPath() string {
	return helpers.GetPlayerPath(group.PlayerID, group.GetPlayerName())
}

func (group PlayerGroup) GetType() string {
	return helpers.GetGroupType(group.GroupType)
}

func (group PlayerGroup) GetPlayerName() string {
	return helpers.GetPlayerName(group.PlayerID, group.PlayerName)
}

func (group PlayerGroup) GetPlayerLink() string {
	return helpers.GetPlayerPath(group.PlayerID, group.PlayerName)
}

func (group PlayerGroup) GetPlayerAvatar() string {
	return helpers.GetPlayerAvatar(group.PlayerAvatar)
}

func (group PlayerGroup) GetPlayerAvatar2() string {
	return helpers.GetPlayerAvatar2(group.PlayerLevel)
}

func (group PlayerGroup) IsOfficial() bool {
	return helpers.IsGroupOfficial(group.GroupType)
}

func (group PlayerGroup) GetURL() string {
	return helpers.GetGroupLink(group.GroupType, group.GroupURL)
}

func (group PlayerGroup) GetName() string {
	return helpers.GetGroupName(group.GroupID, group.GroupName)
}

func (group PlayerGroup) GetGroupIcon() string {
	return helpers.GetGroupIcon(group.GroupIcon)
}

func (group PlayerGroup) GetPlayerCommunityLink() string {
	return helpers.GetPlayerCommunityLink(group.PlayerID, group.PlayerName)
}

func (group PlayerGroup) GetPlayerFlag() string {
	return helpers.GetPlayerFlagPath(group.PlayerCountry)
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

	return getPlayerGroups(offset, limit, filter, sort)
}

func GetGroupPlayers(groupID string, offset int64, order bson.D) (players []PlayerGroup, err error) {

	var filter = bson.D{{"group_id", groupID}}

	return getPlayerGroups(offset, 100, filter, order)
}

func getPlayerGroups(offset int64, limit int64, filter bson.D, sort bson.D) (players []PlayerGroup, err error) {

	cur, ctx, err := Find(CollectionPlayerGroups, offset, limit, sort, filter, nil, nil)
	if err != nil {
		return players, err
	}

	defer func(cur *mongo.Cursor) {
		err = cur.Close(ctx)
		log.Err(err)
	}(cur)

	for cur.Next(ctx) {

		var group PlayerGroup
		err := cur.Decode(&group)
		if err != nil {
			// log.Err(err, group.getKey(), cur.Current.String())
		} else {
			players = append(players, group)
		}
	}

	return players, cur.Err()
}
