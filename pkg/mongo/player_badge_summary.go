package mongo

import (
	"errors"
	"fmt"
	"html/template"
	"strconv"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	. "go.mongodb.org/mongo-driver/bson"
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

func (pbs PlayerBadgeSummary) BSON() interface{} {

	return M{
		"_id":            pbs.ID,
		"players":        pbs.PlayersCount,
		"max_level":      pbs.MaxLevel,
		"max_level_foil": pbs.MaxLevelFoil,
		"leaders":        pbs.Leaders,
		"leaders_foil":   pbs.LeadersFoil,
	}
}

func (pbs PlayerBadgeSummary) GetSpecialLeaders() (ret template.HTML) {

	if len(pbs.Leaders) > 1 {
		return template.HTML(strconv.Itoa(len(pbs.Leaders))) + " joint firsts"
	}

	for playerID, playerName := range pbs.Leaders {
		i, err := strconv.ParseInt(playerID, 10, 64)
		if err != nil {
			log.Err(err)
		} else {
			return "<a href=" + template.HTML(helpers.GetPlayerPath(i, playerName)) + ">" + template.HTML(playerName) + "</a>"
		}
	}

	return "None"
}

func (pbs PlayerBadgeSummary) GetAppLeader(foil bool) (ret template.HTML) {

	var leaders map[string]string
	if foil {
		leaders = pbs.LeadersFoil
	} else {
		leaders = pbs.Leaders
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
			D{{"app_id", 0}, {"badge_id", badge.BadgeID}},
			D{{"badge_level", -1}, {"badge_completion_time", 1}},
			M{"badge_level": 1, "badge_completion_time": 1},
			&topPlayerBadge,
		)

		if err != nil {
			return err
		}

		// Get all players with equal top player badge
		winningBadges, err := getPlayerBadges(
			0,
			0,
			D{{"app_id", 0}, {"badge_id", badge.BadgeID}, {"badge_level", topPlayerBadge.BadgeLevel}, {"badge_completion_time", topPlayerBadge.BadgeCompletionTime}},
			D{{"badge_completion_time", -1}},
			M{"badge_level": 1, "_id": -1, "player_id": 1, "player_name": 1},
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
		summary.PlayersCount, err = CountDocuments(CollectionPlayerBadges, D{{"app_id", 0}, {"badge_id", badge.BadgeID}}, 0)
		if err != nil {
			return err
		}

	} else {

		summary.ID = badge.AppID

		// Get the top player badge, foil=false
		err = FindOne(
			CollectionPlayerBadges,
			D{{"app_id", badge.AppID}, {"badge_id", M{"$gt": 0}}, {"badge_foil", false}},
			D{{"badge_level", -1}, {"badge_completion_time", 1}},
			M{"badge_level": 1, "_id": -1, "player_id": 1, "player_name": 1},
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
			D{{"app_id", badge.AppID}, {"badge_id", M{"$gt": 0}}, {"badge_foil", true}},
			D{{"badge_level", -1}, {"badge_completion_time", 1}},
			M{"badge_level": 1, "_id": -1, "player_id": 1, "player_name": 1},
			&topPlayerBadge,
		)

		if err != nil {
			return err
		}

		summary.LeadersFoil = map[string]string{strconv.FormatInt(topPlayerBadge.PlayerID, 10): topPlayerBadge.PlayerName}
		summary.MaxLevelFoil = topPlayerBadge.BadgeLevel

		// Get number of players with badge
		summary.PlayersCount, err = CountDocuments(CollectionPlayerBadges, D{{"app_id", badge.AppID}, {"badge_id", M{"$gt": 0}}}, 0)
		if err != nil {
			return err
		}
	}

	_, err = ReplaceOne(CollectionPlayerBadgesSummary, D{{"_id", summary.ID}}, summary)
	return err
}
