package mongo

import (
	"errors"
	"fmt"
	"html/template"
	"strconv"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
)

type PlayerBadgeSummary struct {
	ID           int               `bson:"_id"` // App/Badge ID
	PlayersCount int64             `bson:"players"`
	MaxLevel     int               `bson:"max_level"`
	MaxLevelFoil int               `bson:"max_level_foil"`
	Leaders      map[string]string `bson:"leaders"`      // Must use string for playerID as it's a JSON key
	LeadersFoil  map[string]string `bson:"leaders_foil"` // Must use string for playerID as it's a JSON key
	Badge        PlayerBadge       `bson:"-"`
}

func (badge PlayerBadgeSummary) BSON() bson.D {

	return bson.D{
		{"_id", badge.ID},
		{"players", badge.PlayersCount},
		{"max_level", badge.MaxLevel},
		{"max_level_foil", badge.MaxLevelFoil},
		{"leaders", badge.Leaders},
		{"leaders_foil", badge.LeadersFoil},
	}
}

func (badge PlayerBadgeSummary) GetSpecialLeaders() (ret template.HTML) {

	if len(badge.Leaders) > 1 {
		return template.HTML(strconv.Itoa(len(badge.Leaders))) + " joint firsts"
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

	cur, ctx, err := Find(CollectionPlayerBadgesSummary, 0, 0, nil, nil, nil, nil)
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
			log.Err(err, fmt.Sprint(badge))
		}
		badge.Badge = GlobalBadges[badge.ID]
		badges = append(badges, badge)
	}

	return badges, cur.Err()
}

func UpdateBadgeSummary(id int) (err error) {

	var summary PlayerBadgeSummary
	var topPlayerBadge PlayerBadge

	badge, ok := GlobalBadges[id]
	if !ok {
		return errors.New("invalid badge key")
	}

	if badge.IsSpecial() {

		summary.ID = badge.BadgeID

		// Get the top player badge
		err = FindOne(
			CollectionPlayerBadges,
			bson.D{{"app_id", 0}, {"badge_id", badge.BadgeID}},
			bson.D{{"badge_level", -1}, {"badge_completion_time", 1}},
			bson.M{"badge_level": 1, "badge_completion_time": 1},
			&topPlayerBadge,
		)

		if err != nil {
			return err
		}

		// Get all players with equal top player badge
		winningBadges, err := getPlayerBadges(
			0,
			0,
			bson.D{{"app_id", 0}, {"badge_id", badge.BadgeID}, {"badge_level", topPlayerBadge.BadgeLevel}, {"badge_completion_time", topPlayerBadge.BadgeCompletionTime}},
			bson.D{{"badge_completion_time", -1}},
			bson.M{"badge_level": 1, "_id": -1, "player_id": 1, "player_name": 1},
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

	} else {

		summary.ID = badge.AppID

		// Get the top player badge, foil=false
		err = FindOne(
			CollectionPlayerBadges,
			bson.D{{"app_id", badge.AppID}, {"badge_id", bson.M{"$gt": 0}}, {"badge_foil", false}},
			bson.D{{"badge_level", -1}, {"badge_completion_time", 1}},
			bson.M{"badge_level": 1, "_id": -1, "player_id": 1, "player_name": 1},
			&topPlayerBadge,
		)

		if err != nil {
			return err
		}

		summary.Leaders = map[string]string{strconv.FormatInt(topPlayerBadge.PlayerID, 10): topPlayerBadge.PlayerName}
		summary.MaxLevel = topPlayerBadge.BadgeLevel

		// Get the top player badge, foil=true
		err = FindOne(
			CollectionPlayerBadges,
			bson.D{{"app_id", badge.AppID}, {"badge_id", bson.M{"$gt": 0}}, {"badge_foil", true}},
			bson.D{{"badge_level", -1}, {"badge_completion_time", 1}},
			bson.M{"badge_level": 1, "_id": -1, "player_id": 1, "player_name": 1},
			&topPlayerBadge,
		)

		if err != nil {
			return err
		}

		summary.LeadersFoil = map[string]string{strconv.FormatInt(topPlayerBadge.PlayerID, 10): topPlayerBadge.PlayerName}
		summary.MaxLevelFoil = topPlayerBadge.BadgeLevel

		// Get number of players with badge
		summary.PlayersCount, err = CountDocuments(CollectionPlayerBadges, bson.D{{"app_id", badge.AppID}, {"badge_id", bson.M{"$gt": 0}}}, 0)
		if err != nil {
			return err
		}
	}

	_, err = ReplaceOne(CollectionPlayerBadgesSummary, bson.D{{"_id", summary.ID}}, summary)
	return err
}
