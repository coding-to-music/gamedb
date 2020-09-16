package mongo

import (
	"strconv"
	"time"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PlayerBadge struct {
	AppID               int       `bson:"app_id"`
	AppName             string    `bson:"app_name"`
	BadgeCompletionTime time.Time `bson:"badge_completion_time"`
	BadgeFoil           bool      `bson:"badge_foil"`
	BadgeIcon           string    `bson:"badge_icon"`
	BadgeID             int       `bson:"badge_id"`
	BadgeItemID         int64     `bson:"-"`
	BadgeLevel          int       `bson:"badge_level"`
	BadgeScarcity       int       `bson:"badge_scarcity"`
	BadgeXP             int       `bson:"badge_xp"`
	PlayerID            int64     `bson:"player_id"`
	PlayerName          string    `bson:"player_name"`
	PlayerIcon          string    `bson:"player_icon"`
}

func (badge PlayerBadge) BSON() bson.D {

	return bson.D{
		{"_id", badge.GetKey()},
		{"app_id", badge.AppID},
		{"app_name", badge.AppName},
		{"badge_completion_time", badge.BadgeCompletionTime},
		{"badge_foil", badge.BadgeFoil},
		{"badge_icon", badge.BadgeIcon},
		{"badge_id", badge.BadgeID},
		{"badge_level", badge.BadgeLevel},
		{"badge_scarcity", badge.BadgeScarcity},
		{"badge_xp", badge.BadgeXP},
		{"player_id", badge.PlayerID},
		{"player_icon", badge.PlayerIcon},
		{"player_name", badge.PlayerName},
	}
}

func (badge PlayerBadge) GetKey() string {
	return strconv.FormatInt(badge.PlayerID, 10) + "-" + strconv.Itoa(badge.AppID) + "-" + strconv.Itoa(badge.BadgeID) + "-" + strconv.FormatBool(badge.BadgeFoil)
}

func (badge PlayerBadge) IsSpecial() bool {
	return helpers.IsBadgeSpecial(badge.AppID)
}

func (badge PlayerBadge) IsEvent() bool {
	return helpers.IsBadgeEvent(badge.BadgeID)
}

func (badge PlayerBadge) IsGame(appID int) bool {
	return helpers.IsBadgeGame(appID)
}

func (badge PlayerBadge) ID() int {
	return helpers.GetBadgeUniqueID(badge.AppID, badge.BadgeID)
}

func (badge PlayerBadge) GetName() string {
	return helpers.GetBadgeName(badge.AppName, helpers.GetBadgeUniqueID(badge.AppID, badge.BadgeID))
}

func (badge PlayerBadge) GetPath() string {
	return helpers.GetBadgePath(badge.AppName, badge.AppID, badge.BadgeID, badge.BadgeFoil)
}

func (badge PlayerBadge) GetPathToggle() string {
	return helpers.GetBadgePath(badge.AppName, badge.AppID, badge.BadgeID, !badge.BadgeFoil)
}

func (badge PlayerBadge) GetIcon() string {
	return helpers.GetBadgeIcon(badge.BadgeIcon, badge.AppID, badge.BadgeID)
}

func (badge PlayerBadge) GetPlayerIcon() string {
	return helpers.GetPlayerAvatar(badge.PlayerIcon)
}

func (badge PlayerBadge) GetPlayerName() string {
	return helpers.GetPlayerName(badge.PlayerID, badge.PlayerName)
}

func (badge PlayerBadge) GetAppPath() string {
	return helpers.GetAppPath(badge.AppID, helpers.GetAppName(badge.AppID, badge.AppName))
}

func (badge PlayerBadge) GetPlayerPath() string {
	return helpers.GetPlayerPath(badge.PlayerID, badge.PlayerName) + "#badges"
}

// func (badge PlayerBadge) GetPathToggle() string {
//
// 	badge.BadgeFoil = !badge.BadgeFoil
// 	return badge.GetPath()
// }

func (badge PlayerBadge) GetPlayerCommunityLink() string {

	var dir string
	if badge.IsSpecial() {
		dir = "gamecards"
	} else {
		dir = "badges"
	}

	return "https://steamcommunity.com/profiles/" + strconv.FormatInt(badge.PlayerID, 10) + "/" + dir + "/" + strconv.Itoa(badge.BadgeID)
}

func (badge PlayerBadge) GetType() string {

	switch {
	case badge.IsSpecial():
		return "Special"
	case badge.IsEvent():
		return "Event"
	default:
		return "Game"
	}
}

func ReplacePlayerBadges(badges []PlayerBadge) (err error) {

	if len(badges) == 0 {
		return nil
	}

	client, ctx, err := getMongo()
	if err != nil {
		return err
	}

	var writes []mongo.WriteModel
	for _, badge := range badges {

		write := mongo.NewReplaceOneModel()
		write.SetFilter(bson.M{"_id": badge.GetKey()})
		write.SetReplacement(badge.BSON())
		write.SetUpsert(true)

		writes = append(writes, write)
	}

	c := client.Database(config.C.MongoDatabase).Collection(CollectionPlayerBadges.String())

	_, err = c.BulkWrite(ctx, writes, options.BulkWrite())

	return err
}

func GetPlayerBadges(offset int64, filter bson.D, sort bson.D) (badges []PlayerBadge, err error) {
	return getPlayerBadges(offset, 100, filter, sort, nil)
}

// Get the first PlayerBadge for an app ID
func GetAppBadge(appID int) (badge PlayerBadge, err error) {

	var item = memcache.MemcacheFirstAppBadge(appID)

	err = memcache.GetSetInterface(item.Key, item.Expiration, &badge, func() (interface{}, error) {

		badges, err := getPlayerBadges(0, 1, bson.D{{"app_id", appID}}, nil, nil)
		if err != nil {
			return nil, err
		}

		if len(badges) == 0 {
			return nil, ErrNoDocuments
		}

		return badges[0], nil
	})

	return badge, err
}

func GetBadgePlayers(offset int64, filter bson.D) (badges []PlayerBadge, err error) {
	return getPlayerBadges(offset, 100, filter, bson.D{{"badge_level", -1}, {"badge_completion_time", 1}}, nil)
}

func getPlayerBadges(offset int64, limit int64, filter bson.D, sort bson.D, projection bson.M) (badges []PlayerBadge, err error) {

	cur, ctx, err := Find(CollectionPlayerBadges, offset, limit, sort, filter, projection, nil)
	if err != nil {
		return badges, err
	}

	defer close(cur, ctx)

	for cur.Next(ctx) {

		var badge PlayerBadge
		err := cur.Decode(&badge)
		if err != nil {
			log.ErrS(err, badge.GetKey())
		} else {
			badges = append(badges, badge)
		}
	}

	return badges, cur.Err()
}
