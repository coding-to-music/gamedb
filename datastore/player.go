package datastore

import (
	"errors"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/Jleagle/go-helpers/logger"
	"github.com/gosimple/slug"
	"github.com/steam-authority/steam-authority/helpers"
	"github.com/steam-authority/steam-authority/steam"
)

const (
	NotFound = "datastore: no such entity"
)

type Player struct {
	CreatedAt        time.Time                   `datastore:"created_at"`
	UpdatedAt        time.Time                   `datastore:"updated_at"`
	FriendsAddedAt   time.Time                   `datastore:"friends_added_at,noindex"`
	PlayerID         int                         `datastore:"player_id"`
	ValintyURL       string                      `datastore:"vality_url,noindex"`
	Avatar           string                      `datastore:"avatar,noindex"`
	PersonaName      string                      `datastore:"persona_name,noindex"`
	RealName         string                      `datastore:"real_name,noindex"`
	CountryCode      string                      `datastore:"country_code"`
	StateCode        string                      `datastore:"status_code"`
	Level            int                         `datastore:"level"`
	Games            []steam.OwnedGame           `datastore:"games,noindex"`
	GamesRecent      []steam.RecentlyPlayedGame  `datastore:"games_recent,noindex"`
	GamesCount       int                         `datastore:"games_count"`
	Badges           steam.BadgesResponse        `datastore:"badges,noindex"`
	BadgesCount      int                         `datastore:"badges_count"`
	PlayTime         int                         `datastore:"play_time"`
	TimeCreated      int                         `datastore:"time_created"` // In Steam's DB
	Friends          []steam.GetFriendListFriend `datastore:"friends,noindex"`
	FriendsCount     int                         `datastore:"friends_count"`
	Donated          int                         `datastore:"donated"` // Total
	Bans             steam.GetPlayerBanResponse  `datastore:"bans"`
	NumberOfVACBans  int                         `datastore:"bans_cav"`
	NumberOfGameBans int                         `datastore:"bans_game"`
	Groups           []int                       `datastore:"groups"`
}

func (p Player) GetKey() (key *datastore.Key) {
	return datastore.NameKey(KindPlayer, strconv.Itoa(p.PlayerID), nil)
}

func (p Player) GetPath() string {
	return "/players/" + strconv.Itoa(p.PlayerID) + "/" + slug.Make(p.PersonaName)
}

func (p Player) GetAvatar() string {
	if strings.HasPrefix(p.Avatar, "http") {
		return p.Avatar
	} else {
		return "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/" + p.Avatar
	}
}

func (p Player) GetFlag() string {
	return "/assets/img/flags/" + strings.ToLower(p.CountryCode) + ".png"
}

func (p Player) shouldUpdate() bool {

	if p.PersonaName == "" {
		return true
	}

	if p.UpdatedAt.Unix() < (time.Now().Unix() - int64(60*60*24)) {
		return true
	}

	return false
}

func (p Player) ShouldUpdateFriends() bool {
	return p.FriendsAddedAt.Unix() < (time.Now().Unix() - int64(60*60*24*30))
}

func (p Player) GetTimeShort() (ret string) {
	return helpers.GetTimeShort(p.PlayTime, 2)
}

func (p Player) GetTimeLong() (ret string) {
	return helpers.GetTimeLong(p.PlayTime, 5)
}

func GetPlayer(id int) (ret *Player, err error) {

	client, context, err := getDSClient()
	if err != nil {
		return ret, err
	}

	key := datastore.NameKey(KindPlayer, strconv.Itoa(id), nil)

	player := new(Player)
	player.PlayerID = id

	err = client.Get(context, key, player)
	if err != nil {

		if err.Error() == NotFound {
			return player, nil
		}
		return player, err
	}

	return player, nil
}

func GetPlayers(order string, limit int) (players []*Player, err error) {

	client, ctx, err := getDSClient()
	if err != nil {
		return players, err
	}

	q := datastore.NewQuery(KindPlayer).Order(order).Limit(limit)
	_, err = client.GetAll(ctx, q, &players)

	return players, err
}

func GetPlayersByIDs(ids []int) (friends []Player, err error) {

	if len(ids) > 1000 {
		return friends, errors.New("too many")
	}

	client, context, err := getDSClient()
	if err != nil {
		return friends, err
	}

	var keys []*datastore.Key
	for _, v := range ids {
		key := datastore.NameKey(KindPlayer, strconv.Itoa(v), nil)
		keys = append(keys, key)
	}

	friends = make([]Player, len(keys))
	err = client.GetMulti(context, keys, friends)
	if err != nil && !strings.Contains(err.Error(), "no such entity") {
		return friends, err
	}

	// Sort friends by level desc
	sort.Slice(friends, func(i, j int) bool {
		return friends[i].Level > friends[j].Level
	})

	return friends, nil
}

func CountPlayers() (count int, err error) {

	client, ctx, err := getDSClient()
	if err != nil {
		return count, err
	}

	q := datastore.NewQuery(KindPlayer)
	count, err = client.Count(ctx, q)
	if err != nil {
		return count, err
	}

	return count, nil
}

func (p *Player) UpdateIfNeeded() (err error) {

	if p.shouldUpdate() {

		err = p.fill()
		if err != nil {
			return err
		}

		err = p.Save()
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Player) fill() (err error) {

	//Get summary
	summary, err := steam.GetPlayerSummaries(p.PlayerID)
	if err != nil {
		if err.Error() == steam.ErrorInvalidJson {
			return err
		}
		if !strings.HasPrefix(err.Error(), "not found in steam") {
			logger.Error(err)
		}
	}

	p.Avatar = strings.Replace(summary.AvatarFull, "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/", "", 1)
	p.ValintyURL = path.Base(summary.ProfileURL)
	p.RealName = summary.RealName
	p.CountryCode = summary.LOCCountryCode
	p.StateCode = summary.LOCStateCode
	p.PersonaName = summary.PersonaName

	// Get games
	gamesResponse, err := steam.GetOwnedGames(p.PlayerID)
	if err != nil {
		if err.Error() == steam.ErrorInvalidJson {
			return err
		}
		logger.Error(err)
	}

	p.Games = gamesResponse
	p.GamesCount = len(gamesResponse)

	// Get playtime
	var playtime = 0
	for _, v := range gamesResponse {
		playtime = playtime + v.PlaytimeForever
	}
	p.PlayTime = playtime

	// Get recent games
	recentGames, err := steam.GetRecentlyPlayedGames(p.PlayerID)
	if err != nil {
		if err.Error() == steam.ErrorInvalidJson {
			return err
		}
		logger.Error(err)
	}

	p.GamesRecent = recentGames

	// Get badges
	badges, err := steam.GetBadges(p.PlayerID)
	if err != nil {
		if err.Error() == steam.ErrorInvalidJson {
			return err
		}
		logger.Error(err)
	}

	p.Badges = badges
	p.BadgesCount = len(badges.Badges)

	//Get friends
	friends, err := steam.GetFriendList(p.PlayerID)
	if err != nil {
		if err.Error() == steam.ErrorInvalidJson {
			return err
		}
		logger.Error(err)
	}

	p.Friends = friends
	p.FriendsCount = len(friends)

	// Get level
	level, err := steam.GetSteamLevel(p.PlayerID)
	if err != nil {
		if err.Error() == steam.ErrorInvalidJson {
			return err
		}
		logger.Error(err)
	}

	p.Level = level

	// Get bans
	bans, err := steam.GetPlayerBans(p.PlayerID)
	if err != nil {
		if err.Error() == steam.ErrorInvalidJson {
			return err
		}
		logger.Error(err)
	}

	p.Bans = bans
	p.NumberOfGameBans = bans.NumberOfGameBans
	p.NumberOfVACBans = bans.NumberOfVACBans

	// Get groups
	groups, err := steam.GetUserGroupList(p.PlayerID)
	if err != nil {
		if err.Error() == steam.ErrorInvalidJson {
			return err
		}
		logger.Error(err)
	}

	p.Groups = groups

	// Fix dates
	p.UpdatedAt = time.Now()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now()
	}

	return nil
}

func (p *Player) Save() (err error) {

	if p.PlayerID == 0 {
		logger.Info("Saving player with ID 0")
	}

	_, err = SaveKind(p.GetKey(), p)
	if err != nil {
		return err
	}

	return nil
}
