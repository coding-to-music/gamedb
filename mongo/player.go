package mongo

import (
	"time"

	"github.com/gamedb/website/helpers"
	"go.mongodb.org/mongo-driver/bson"
)

type Player struct {
	ID               int64     `bson:"_id"` //
	Avatar           string    ``           //
	Badges           string    ``           // []ProfileBadge
	BadgesCount      int       ``           //
	BadgeStats       string    ``           // ProfileBadgeStats
	Bans             string    ``           // PlayerBans
	CountryCode      string    ``           //
	Donated          int       ``           //
	Friends          string    ``           // []ProfileFriend
	FriendsAddedAt   time.Time ``           // delete?
	FriendsCount     int       ``           //
	GamesCount       int       ``           //
	GamesRecent      string    ``           // []ProfileRecentGame
	GameStats        string    ``           // PlayerAppStatsTemplate
	Groups           []int     ``           // []int
	LastLogOff       time.Time ``           //
	Level            int       ``           //
	NumberOfGameBans int       ``           //
	NumberOfVACBans  int       ``           //
	PersonaName      string    ``           //
	PlayTime         int       ``           //
	PrimaryClanID    int       ``           //
	RealName         string    ``           //
	StateCode        string    ``           //
	TimeCreated      time.Time ``           //
	UpdatedAt        time.Time ``           //
	VanintyURL       string    ``           //
}

func (player Player) Key() interface{} {
	return player.ID
}

func (player Player) BSON() (ret interface{}) {

	return bson.M{
		"_id":              player.ID,
		"avatar":           player.Avatar,
		"badges":           player.Badges,
		"badges_count":     player.BadgesCount,
		"badge_stats":      player.BadgeStats,
		"bans":             player.Bans,
		"country_code":     player.CountryCode,
		"donated":          player.Donated,
		"friends":          player.Friends,
		"friends_added_at": player.FriendsAddedAt,
		"friends_count":    player.FriendsCount,
		"games_count":      player.GamesCount,
		"games_recent":     player.GamesRecent,
		"game_stats":       player.GameStats,
		"groups":           player.Groups,
		"time_logged_off":  player.LastLogOff,
		"level":            player.Level,
		"bans_game":        player.NumberOfGameBans,
		"bans_cav":         player.NumberOfVACBans,
		"persona_name":     player.PersonaName,
		"play_time":        player.PlayTime,
		"primary_clan_id":  player.PrimaryClanID,
		"real_name":        player.RealName,
		"status_code":      player.StateCode,
		"time_created":     player.TimeCreated,
		"updated_at":       time.Now(),
		"vanity_url":       player.VanintyURL,
	}
}

func CountPlayers() (count int64, err error) {

	var item = helpers.MemcacheCountPlayers

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &count, func() (interface{}, error) {

		return CountDocuments(CollectionPlayers, bson.M{})
	})

	return count, err
}
