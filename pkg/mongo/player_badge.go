package mongo

import (
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PlayerBadge struct {
	PlayerID            int64     `bson:"player_id"`
	BadgeID             int       `bson:"badge_id"`
	BadgeLevel          int       `bson:"badge_level"`
	BadgeCompletionTime time.Time `bson:"badge_time"`
	BadgeXP             int       `bson:"badge_xp"`
	BadgeScarcity       int       `bson:"badge_scarcity"`
	BadgeItemID         int64     `bson:"badge_item_id"`
	AppID               int       `bson:"app_id"`
	AppName             string    `bson:"app_name"`
	AppIcon             string    `bson:"app_icon"`
}

func (pb PlayerBadge) BSON() (ret interface{}) {

	return M{
		"_id":            pb.getKey(),
		"player_id":      pb.PlayerID,
		"badge_id":       pb.BadgeID,
		"badge_level":    pb.BadgeLevel,
		"badge_time":     pb.BadgeCompletionTime,
		"badge_xp":       pb.BadgeXP,
		"badge_scarcity": pb.BadgeScarcity,
		"badge_item_id":  pb.BadgeItemID,
		"app_id":         pb.AppID,
		"app_name":       pb.AppName,
		"app_icon":       pb.AppIcon,
	}
}

func (pb PlayerBadge) getKey() string {
	return strconv.FormatInt(pb.PlayerID, 10) + "-" + strconv.Itoa(pb.AppID) + "-" + strconv.Itoa(pb.BadgeID)
}

func UpdatePlayerBadges(badges []PlayerBadge) (err error) {

	if badges == nil || len(badges) == 0 {
		return nil
	}

	client, ctx, err := getMongo()
	if err != nil {
		return err
	}

	var writes []mongo.WriteModel
	for _, badge := range badges {

		write := mongo.NewReplaceOneModel()
		write.SetFilter(M{"_id": badge.getKey()})
		write.SetReplacement(badge.BSON())
		write.SetUpsert(true)

		writes = append(writes, write)
	}

	c := client.Database(MongoDatabase).Collection(CollectionPlayerBadges.String())

	_, err = c.BulkWrite(ctx, writes, options.BulkWrite())

	return err
}
