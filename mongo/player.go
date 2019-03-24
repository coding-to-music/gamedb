package mongo

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const defaultPlayerAvatar = "/assets/img/no-player-image.jpg"

var (
	ErrInvalidPlayerID   = errors.New("invalid id")
	ErrInvalidPlayerName = errors.New("invalid name")
)

type Player struct {
	ID               int64     `bson:"_id"`             //
	Avatar           string    ``                       //
	Badges           string    ``                       // []ProfileBadge
	BadgesCount      int       `bson:"badges_count"`    //
	BadgeStats       string    `bson:"badge_stats"`     // ProfileBadgeStats
	Bans             string    ``                       // PlayerBans
	CountryCode      string    `bson:"country_code"`    //
	Donated          int       ``                       //
	Friends          string    ``                       // []ProfileFriend
	FriendsCount     int       `bson:"friends_count"`   //
	GamesCount       int       `bson:"games_count"`     //
	GamesRecent      string    `bson:"games_recent"`    // []ProfileRecentGame
	GameStats        string    `bson:"game_stats"`      // PlayerAppStatsTemplate
	Groups           []int     ``                       // []int
	LastLogOff       time.Time `bson:"time_logged_off"` //
	Level            int       ``                       //
	NumberOfGameBans int       `bson:"bans_game"`       //
	NumberOfVACBans  int       `bson:"bans_cav"`        //
	PersonaName      string    `bson:"persona_name"`    //
	PlayTime         int       `bson:"play_time"`       //
	PrimaryClanID    int       `bson:"primary_clan_id"` //
	RealName         string    `bson:"real_name"`       //
	StateCode        string    `bson:"status_code"`     //
	TimeCreated      time.Time `bson:"time_created"`    //
	UpdatedAt        time.Time `bson:"updated_at"`      //
	VanintyURL       string    `bson:"vanity_url"`      //
}

func (player Player) Key() interface{} {
	return player.ID
}

func (player Player) BSON() (ret interface{}) {

	return bson.M{
		"_id":             player.ID,
		"avatar":          player.Avatar,
		"badges":          player.Badges,
		"badges_count":    player.BadgesCount,
		"badge_stats":     player.BadgeStats,
		"bans":            player.Bans,
		"country_code":    player.CountryCode,
		"donated":         player.Donated,
		"friends":         player.Friends,
		"friends_count":   player.FriendsCount,
		"games_count":     player.GamesCount,
		"games_recent":    player.GamesRecent,
		"game_stats":      player.GameStats,
		"groups":          player.Groups,
		"time_logged_off": player.LastLogOff,
		"level":           player.Level,
		"bans_game":       player.NumberOfGameBans,
		"bans_cav":        player.NumberOfVACBans,
		"persona_name":    player.PersonaName,
		"play_time":       player.PlayTime,
		"primary_clan_id": player.PrimaryClanID,
		"real_name":       player.RealName,
		"status_code":     player.StateCode,
		"time_created":    player.TimeCreated,
		"updated_at":      time.Now(),
		"vanity_url":      player.VanintyURL,
	}
}

func (player Player) GetPath() string {
	return helpers.GetPlayerPath(player.ID, player.GetName())
}

func (player Player) GetName() string {
	return helpers.GetPlayerName(player.ID, player.PersonaName)
}

func (player Player) GetSteamTimeUnix() int64 {
	return player.TimeCreated.Unix()
}

func (player Player) GetSteamTimeNice() string {
	return player.TimeCreated.Format(helpers.DateYear)
}

func (player Player) GetLogoffUnix() int64 {
	return player.LastLogOff.Unix()
}

func (player Player) GetLogoffNice() string {
	return player.LastLogOff.Format(helpers.DateYearTime)
}

func (player Player) GetUpdatedUnix() int64 {
	return player.UpdatedAt.Unix()
}

func (player Player) GetUpdatedNice() string {
	return player.UpdatedAt.Format(helpers.DateTime)
}

func (player Player) GetSteamCommunityLink() string {
	return "https://steamcommunity.com/profiles/" + strconv.FormatInt(player.ID, 10)
}

func (player Player) GetMaxFriends() int {
	return helpers.GetPlayerMaxFriends(player.Level)
}

func (player Player) GetAvatar() string {
	if strings.HasPrefix(player.Avatar, "http") {
		return player.Avatar
	} else if player.Avatar != "" {
		return "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/" + player.Avatar
	} else {
		return player.GetDefaultAvatar()
	}
}

func (player Player) GetDefaultAvatar() string {
	return defaultPlayerAvatar
}

func (player Player) GetFlag() string {
	return "/assets/img/flags/" + strings.ToLower(player.CountryCode) + ".png"
}

func (player Player) GetCountry() string {
	return helpers.CountryCodeToName(player.CountryCode)
}

func (player Player) GetBadgeStats() (stats ProfileBadgeStats, err error) {

	err = helpers.Unmarshal([]byte(player.BadgeStats), &stats)
	return stats, err
}

func (player Player) GetBadges() (badges []ProfileBadge, err error) {

	if player.Badges == "" || player.Badges == "null" {
		return
	}

	var bytes []byte

	if helpers.IsStorageLocaion(player.Badges) {

		bytes, err = helpers.Download(helpers.PathBadges(player.ID))
		err = helpers.IgnoreErrors(err, storage.ErrObjectNotExist)
		if err != nil {
			return badges, err
		}
	} else {
		bytes = []byte(player.Badges)
	}

	err = helpers.Unmarshal(bytes, &badges)
	return badges, err
}

func (player Player) GetFriends() (friends []ProfileFriend, err error) {

	if player.Friends == "" || player.Friends == "null" {
		return
	}

	var bytes []byte

	if helpers.IsStorageLocaion(player.Friends) {

		bytes, err = helpers.Download(helpers.PathFriends(player.ID))
		err = helpers.IgnoreErrors(err, storage.ErrObjectNotExist)
		if err != nil {
			return friends, err
		}
	} else {
		bytes = []byte(player.Friends)
	}

	err = helpers.Unmarshal(bytes, &friends)
	return friends, err
}

func (player Player) GetRecentGames() (games []ProfileRecentGame, err error) {

	if player.GamesRecent == "" || player.GamesRecent == "null" {
		return
	}

	var bytes []byte

	if helpers.IsStorageLocaion(player.GamesRecent) {

		bytes, err = helpers.Download(helpers.PathRecentGames(player.ID))
		err = helpers.IgnoreErrors(err, storage.ErrObjectNotExist)
		if err != nil {
			return games, err
		}
	} else {
		bytes = []byte(player.GamesRecent)
	}

	err = helpers.Unmarshal(bytes, &games)
	return games, err
}

func (player Player) GetBans() (bans PlayerBans, err error) {

	err = helpers.Unmarshal([]byte(player.Bans), &bans)
	return bans, err
}

func (player Player) GetGameStats(code steam.CountryCode) (stats PlayerAppStatsTemplate, err error) {

	err = helpers.Unmarshal([]byte(player.GameStats), &stats)

	stats.All.Code = code
	stats.Played.Code = code

	return stats, err
}

type UpdateType string

const (
	PlayerUpdateAuto   UpdateType = "auto"
	PlayerUpdateManual UpdateType = "manual"
	PlayerUpdateAdmin  UpdateType = "admin"
)

func (player Player) ShouldUpdate(userAgent string, updateType UpdateType) bool {

	if !IsValidPlayerID(player.ID) {
		return false
	}

	if helpers.IsBot(userAgent) {
		return false
	}

	switch updateType {
	case PlayerUpdateAdmin:
		return true
	case PlayerUpdateAuto:
		if player.UpdatedAt.Add(time.Hour * 24 * 7).Unix() < time.Now().Unix() { // 1 week
			return true
		}
	case PlayerUpdateManual:
		if player.Donated == 0 {
			if player.UpdatedAt.Add(time.Hour * 24).Unix() < time.Now().Unix() { // 1 day
				return true
			}
		} else {
			if player.UpdatedAt.Add(time.Hour * 1).Unix() < time.Now().Unix() { // 1 hour
				return true
			}
		}
	}

	return false
}

func GetPlayer(id int64) (player Player, err error) {

	if !IsValidPlayerID(id) {
		return player, ErrInvalidPlayerID
	}

	err = FindDocument(CollectionPlayers, "_id", id, &player)
	return player, err
}

func GetPlayersByIDs(ids []int64) (players []Player, err error) {

	if len(ids) < 1 {
		return players, nil
	}

	client, ctx, err := GetMongo()
	if err != nil {
		return players, err
	}

	var idsBSON bson.A
	for _, v := range ids {
		idsBSON = append(idsBSON, v)
	}

	c := client.Database(MongoDatabase).Collection(CollectionPlayers)
	cur, err := c.Find(ctx, bson.M{"_id": bson.M{"$in": idsBSON}}, options.Find())
	if err != nil {
		return players, err
	}

	defer func() {
		err = cur.Close(ctx)
		log.Err(err)
	}()

	for cur.Next(ctx) {

		var player Player
		err := cur.Decode(&player)
		log.Err(err)
		players = append(players, player)
	}

	return players, err
}

func CountPlayers() (count int64, err error) {

	var item = helpers.MemcachePlayersCount

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &count, func() (interface{}, error) {

		return CountDocuments(CollectionPlayers, bson.M{})
	})

	return count, err
}

func IsValidPlayerID(id int64) bool {

	if id == 0 {
		return false
	}

	idString := strconv.FormatInt(id, 10)

	if !strings.HasPrefix(idString, "76") {
		return false
	}

	if len(idString) != 17 {
		return false
	}

	return true
}
