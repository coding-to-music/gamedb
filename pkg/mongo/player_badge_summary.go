package mongo

import (
	"html/template"
	"strconv"
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
)

type PlayerBadgeSummary struct {
	AppID        int               `bson:"app_id"`
	BadgeID      int               `bson:"badge_id"`
	AppName      string            `bson:"app_name"`
	AppIcon      string            `bson:"app_icon"`
	PlayersCount int64             `bson:"players"`
	MaxLevel     int               `bson:"max_level"`
	MaxLevelFoil int               `bson:"max_level_foil"`
	Leaders      map[string]string `bson:"leaders"`      // Must use string for playerID as it's a JSON key
	LeadersFoil  map[string]string `bson:"leaders_foil"` // Must use string for playerID as it's a JSON key
	UpdatedAt    int64             `bson:"updated_at"`
}

func (badge PlayerBadgeSummary) BSON() bson.D {

	return bson.D{
		{"_id", badge.getKey()},
		{"app_id", badge.AppID},
		{"badge_id", badge.BadgeID},
		{"app_name", badge.AppName},
		{"app_icon", badge.AppIcon},
		{"players", badge.PlayersCount},
		{"max_level", badge.MaxLevel},
		{"max_level_foil", badge.MaxLevelFoil},
		{"leaders", badge.Leaders},
		{"leaders_foil", badge.LeadersFoil},
		{"updated_at", time.Now().Unix()},
	}
}

func (badge PlayerBadgeSummary) getKey() string {
	return strconv.Itoa(badge.AppID) + "-" + strconv.Itoa(badge.BadgeID)
}

func (badge PlayerBadgeSummary) IsSpecial() bool {
	return helpers.IsBadgeSpecial(badge.AppID)
}

func (badge PlayerBadgeSummary) IsEvent() bool {
	return helpers.IsBadgeEvent(badge.AppID)
}

func (badge PlayerBadgeSummary) IsGame() bool {
	return helpers.IsBadgeGame(badge.AppID)
}

func (badge PlayerBadgeSummary) GetPath(foil bool) string {
	return helpers.GetBadgePath(badge.AppName, badge.AppID, badge.BadgeID, foil)
}

func (badge PlayerBadgeSummary) GetName() string {
	return helpers.GetBadgeName(badge.AppName, helpers.GetBadgeUniqueID(badge.AppID, badge.BadgeID))
}

func (badge PlayerBadgeSummary) ID() int {
	return helpers.GetBadgeUniqueID(badge.AppID, badge.BadgeID)
}

func (badge PlayerBadgeSummary) GetIcon() string {
	return helpers.GetBadgeIcon(badge.AppIcon, badge.AppID, badge.BadgeID)
}

func (badge PlayerBadgeSummary) GetSpecialLeaders() (ret template.HTML) {

	if len(badge.Leaders) > 1 {
		return template.HTML(strconv.Itoa(len(badge.Leaders)) + " joint firsts")
	}

	for playerID, playerName := range badge.Leaders {
		i, err := strconv.ParseInt(playerID, 10, 64)
		if err != nil {
			log.Err(err)
		} else {
			return "<a href=" + template.HTML(helpers.GetPlayerPath(i, playerName)) + ">" + template.HTML(playerName) + "</a>"
		}
	}

	return "None"
}

func (badge PlayerBadgeSummary) GetAppLeader(foil bool) (ret template.HTML) {

	var leaders map[string]string
	if foil {
		leaders = badge.LeadersFoil
	} else {
		leaders = badge.Leaders
	}

	for playerID, playerName := range leaders {
		i, err := strconv.ParseInt(playerID, 10, 64)
		if err != nil {
			log.Err(err)
		} else {
			return "<a href=" + template.HTML(helpers.GetPlayerPath(i, playerName)) + ">" + template.HTML(playerName) + "</a>"
		}
	}

	return "None"
}

func GetBadgeSummaries() (badges []PlayerBadgeSummary, err error) {

	// Filter to remove game badges that are no longer popular
	filter := bson.D{{"updated_at", bson.M{"$gte": time.Now().Add(time.Hour * 24 * 7 * -1).Unix()}}}

	cur, ctx, err := Find(CollectionPlayerBadgesSummary, 0, 0, nil, filter, nil, nil)
	if err != nil {
		return badges, err
	}

	defer func() {
		err = cur.Close(ctx)
		log.Err(err)
	}()

	for cur.Next(ctx) {

		var badge PlayerBadgeSummary
		err := cur.Decode(&badge)
		if err != nil {
			log.Err(err, badge)
			continue
		}

		badges = append(badges, badge)
	}

	return badges, cur.Err()
}

func UpdateBadgeSummary(id int) (err error) {

	var summary PlayerBadgeSummary
	var topPlayerBadge PlayerBadge

	//
	badge, ok := helpers.BuiltInSpecialBadges[id]
	if ok {
		summary.AppName = badge.GetName()
		summary.AppIcon = badge.GetIcon()
	} else {
		badge, ok = helpers.BuiltInEventBadges[id]
		if ok {
			summary.AppName = badge.GetName()
			summary.AppIcon = badge.GetIcon()
		} else {

			app, err := GetApp(id)
			if err != nil {
				return err
			}

			summary.AppName = app.GetName()
			summary.AppIcon = app.GetIcon()

			badge.BadgeID = 1
			badge.AppID = id
		}
	}

	summary.AppID = badge.AppID
	summary.BadgeID = badge.BadgeID

	if badge.IsSpecial() {

		// Get the top player badge
		err = FindOne(
			CollectionPlayerBadges,
			bson.D{{"app_id", 0}, {"badge_id", badge.BadgeID}},
			bson.D{{"badge_level", -1}, {"badge_completion_time", 1}},
			nil,
			&topPlayerBadge,
		)

		if err != nil && err != ErrNoDocuments {
			return err
		} else if err == nil {

			// Get all players with equal top player badge
			winningBadges, err := getPlayerBadges(
				0,
				0,
				bson.D{{"app_id", 0}, {"badge_id", badge.BadgeID}, {"badge_level", topPlayerBadge.BadgeLevel}, {"badge_completion_time", topPlayerBadge.BadgeCompletionTime}},
				bson.D{{"badge_completion_time", -1}},
				nil,
			)

			if err != nil {
				return err
			}

			summary.Leaders = map[string]string{}
			for _, v := range winningBadges {
				s := strconv.FormatInt(v.PlayerID, 10)
				summary.Leaders[s] = v.PlayerName
			}

			// Get number of players with badge
			summary.PlayersCount, err = CountDocuments(CollectionPlayerBadges, bson.D{{"app_id", 0}, {"badge_id", badge.BadgeID}}, 0)
			if err != nil {
				return err
			}
		}

	} else {

		// Get the top player badge, foil=false
		err = FindOne(
			CollectionPlayerBadges,
			bson.D{{"app_id", badge.AppID}, {"badge_id", bson.M{"$gt": 0}}, {"badge_foil", false}},
			bson.D{{"badge_level", -1}, {"badge_completion_time", 1}},
			nil,
			&topPlayerBadge,
		)

		if err != nil && err != ErrNoDocuments {
			return err
		} else if err == nil {
			summary.Leaders = map[string]string{strconv.FormatInt(topPlayerBadge.PlayerID, 10): topPlayerBadge.PlayerName}
			summary.MaxLevel = topPlayerBadge.BadgeLevel
		}

		// Get the top player badge, foil=true
		err = FindOne(
			CollectionPlayerBadges,
			bson.D{{"app_id", badge.AppID}, {"badge_id", bson.M{"$gt": 0}}, {"badge_foil", true}},
			bson.D{{"badge_level", -1}, {"badge_completion_time", 1}},
			nil,
			&topPlayerBadge,
		)

		if err != nil && err != ErrNoDocuments {
			return err
		} else if err == nil {
			summary.LeadersFoil = map[string]string{strconv.FormatInt(topPlayerBadge.PlayerID, 10): topPlayerBadge.PlayerName}
			summary.MaxLevelFoil = topPlayerBadge.BadgeLevel
		}

		// Get number of players with badge
		summary.PlayersCount, err = CountDocuments(CollectionPlayerBadges, bson.D{{"app_id", badge.AppID}, {"badge_id", bson.M{"$gt": 0}}}, 0)
		if err != nil {
			return err
		}
	}

	_, err = ReplaceOne(CollectionPlayerBadgesSummary, bson.D{{"_id", summary.getKey()}}, summary)
	return err
}
