package datastore

import (
	"encoding/json"
	"errors"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/Jleagle/steam-go/steam"
	"github.com/gosimple/slug"
	"github.com/steam-authority/steam-authority/helpers"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/steami"
	"github.com/steam-authority/steam-authority/storage"
)

const maxBytesToStore = 1024 * 10

var (
	ErrInvalidID   = errors.New("invalid id")
	ErrInvalidName = errors.New("invalid name")
)

var (
	cachePlayersCount int
)

type Player struct {
	CreatedAt        time.Time `datastore:"created_at"`               //
	UpdatedAt        time.Time `datastore:"updated_at"`               //
	FriendsAddedAt   time.Time `datastore:"friends_added_at,noindex"` //
	PlayerID         int64     `datastore:"player_id"`                //
	VanintyURL       string    `datastore:"vanity_url"`               //
	Avatar           string    `datastore:"avatar,noindex"`           //
	PersonaName      string    `datastore:"persona_name,noindex"`     //
	RealName         string    `datastore:"real_name,noindex"`        //
	CountryCode      string    `datastore:"country_code"`             //
	StateCode        string    `datastore:"status_code"`              //
	Level            int       `datastore:"level"`                    //
	Games            string    `datastore:"games,noindex"`            // JSON
	GamesRecent      string    `datastore:"games_recent,noindex"`     // JSON
	GamesCount       int       `datastore:"games_count"`              //
	Badges           string    `datastore:"badges,noindex"`           // JSON
	BadgesCount      int       `datastore:"badges_count"`             //
	PlayTime         int       `datastore:"play_time"`                //
	TimeCreated      time.Time `datastore:"time_created"`             //
	LastLogOff       time.Time `datastore:"time_logged_off,noindex"`  //
	PrimaryClanID    int       `datastore:"primary_clan_id,noindex"`  //
	Friends          string    `datastore:"friends,noindex"`          // JSON
	FriendsCount     int       `datastore:"friends_count"`            //
	Donated          int       `datastore:"donated"`                  //
	Bans             string    `datastore:"bans"`                     // JSON
	NumberOfVACBans  int       `datastore:"bans_cav"`                 //
	NumberOfGameBans int       `datastore:"bans_game"`                //
	Groups           []int     `datastore:"groups,noindex"`           //
}

func (p Player) GetKey() (key *datastore.Key) {
	return datastore.NameKey(KindPlayer, strconv.FormatInt(p.PlayerID, 10), nil)
}

func (p Player) GetPath() string {

	x := "/players/" + strconv.FormatInt(p.PlayerID, 10)
	if p.PersonaName != "" {
		x = x + "/" + slug.Make(p.PersonaName)
	}
	return x
}

func (p Player) GetName() string {
	return p.PersonaName
}

func (p Player) GetSteamTimeUnix() (int64) {
	return p.TimeCreated.Unix()
}

func (p Player) GetSteamTimeNice() (string) {
	return p.TimeCreated.Format(helpers.DateYear)
}

func (p Player) GetLogoffUnix() (int64) {
	return p.LastLogOff.Unix()
}

func (p Player) GetLogoffNice() (string) {
	return p.LastLogOff.Format(helpers.DateYearTime)
}

func (p Player) GetUpdatedUnix() (int64) {
	return p.UpdatedAt.Unix()
}

func (p Player) GetUpdatedNice() (string) {
	return p.UpdatedAt.Format(helpers.DateTime)
}

func (p Player) GetSteamCommunityLink() string {
	return "http://steamcommunity.com/profiles/" + strconv.FormatInt(p.PlayerID, 10)
}

func (p Player) GetAvatar() string {
	if strings.HasPrefix(p.Avatar, "http") {
		return p.Avatar
	} else if p.Avatar != "" {
		return "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/" + p.Avatar
	} else {
		return p.GetDefaultAvatar()
	}
}

func (p Player) GetDefaultAvatar() string {
	return "/assets/img/no-player-image.jpg"
}

func (p Player) GetFlag() string {
	return "/assets/img/flags/" + strings.ToLower(p.CountryCode) + ".png"
}

func (p Player) GetCountry() string {
	return helpers.CountryCodeToName(p.CountryCode)
}

func (p Player) GetGames() (games []steam.OwnedGame, err error) {

	if len(p.Games) > 0 {

		var bytes []byte

		if strings.HasPrefix(p.Games, "/") {
			bytes, err = storage.Download(storage.PathGames(p.PlayerID))
			if err != nil {
				return games, err
			}
		} else {
			bytes = []byte(p.Games)
		}

		err = json.Unmarshal(bytes, &games)
		if err != nil {
			if strings.Contains(err.Error(), "cannot unmarshal") {
				logger.Info(err.Error() + " - " + string(bytes))
			} else {
				logger.Error(err)
			}
			return games, err
		}
	}

	return games, nil
}

func (p Player) GetBadges() (badges steam.BadgesResponse, err error) {

	var bytes []byte

	if strings.HasPrefix(p.Badges, "/") {
		bytes, err = storage.Download(storage.PathBadges(p.PlayerID))
		if err != nil {
			return badges, err
		}
	} else {
		bytes = []byte(p.Badges)
	}

	if len(bytes) > 0 {
		err = json.Unmarshal(bytes, &badges)
		if err != nil {
			if strings.Contains(err.Error(), "cannot unmarshal") {
				logger.Info(err.Error() + " - " + string(bytes))
			} else {
				logger.Error(err)
			}
			return badges, err
		}
	}

	return badges, nil
}

func (p Player) GetFriends() (friends steam.FriendsList, err error) {

	var bytes []byte

	if strings.HasPrefix(p.Friends, "/") {
		bytes, err = storage.Download(storage.PathFriends(p.PlayerID))
		if err != nil {
			return friends, err
		}
	} else {
		bytes = []byte(p.Friends)
	}

	if len(bytes) > 0 {
		err = json.Unmarshal(bytes, &friends)
		if err != nil {
			if strings.Contains(err.Error(), "cannot unmarshal") {
				logger.Info(err.Error() + " - " + string(bytes))
			} else {
				logger.Error(err)
			}
			return friends, err
		}
	}

	return friends, nil
}

func (p Player) GetRecentGames() (games []steam.RecentlyPlayedGame, err error) {

	if p.GamesRecent == "" {
		return
	}

	var bytes []byte

	if strings.HasPrefix(p.GamesRecent, "/") {
		bytes, err = storage.Download(storage.PathRecentGames(p.PlayerID))
		if err != nil {
			return games, err
		}
	} else {
		bytes = []byte(p.GamesRecent)
	}

	err = json.Unmarshal(bytes, &games)
	if err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			logger.Info(err.Error() + " - " + string(bytes))
		} else {
			logger.Error(err)
		}
		return games, err
	}

	return games, nil
}

func (p Player) GetBans() (bans steam.GetPlayerBanResponse, err error) {

	bytes := []byte(p.Bans)

	if len(bytes) > 0 {
		err = json.Unmarshal(bytes, &bans)
		if err != nil {
			if strings.Contains(err.Error(), "cannot unmarshal") {
				logger.Info(err.Error() + " - " + string(bytes))
			} else {
				logger.Error(err)
			}
			return bans, err
		}
	}

	return bans, nil
}

// todo, check this is acurate
func IsValidPlayerID(id int64) bool {

	if id < 10000000000000000 {
		return false
	}

	if len(strconv.FormatInt(id, 10)) != 17 {
		return false
	}

	return true
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

func GetPlayer(id int64) (ret Player, err error) {

	if !IsValidPlayerID(id) {
		return ret, ErrInvalidID
	}

	client, ctx, err := getClient()
	if err != nil {
		return ret, err
	}

	key := datastore.NameKey(KindPlayer, strconv.FormatInt(id, 10), nil)

	player := Player{}
	player.PlayerID = id

	err = client.Get(ctx, key, &player)
	if err != nil {
		if err2, ok := err.(*datastore.ErrFieldMismatch); ok {

			old := []string{
				"settings_email",
				"settings_password",
				"settings_alerts",
				"settings_hidden",
			}

			if !helpers.SliceHasString(old, err2.FieldName) {
				return player, err2
			}
		} else {
			return player, err
		}
	}

	return player, nil
}

func GetPlayerByName(name string) (ret Player, err error) {

	if len(name) == 0 {
		return ret, ErrInvalidName
	}

	client, ctx, err := getClient()
	if err != nil {
		return ret, err
	}

	q := datastore.NewQuery(KindPlayer).Filter("vanity_url =", name).Limit(1)

	var players []Player

	_, err = client.GetAll(ctx, q, &players)
	if err != nil {
		return ret, err
	}

	if len(players) > 0 {
		return players[0], nil // Success
	}

	return ret, datastore.ErrNoSuchEntity
}

func GetPlayersByEmail(email string) (ret []Player, err error) {

	client, ctx, err := getClient()
	if err != nil {
		return ret, err
	}

	q := datastore.NewQuery(KindPlayer).Filter("settings_email =", email).Limit(1)

	var players []Player

	_, err = client.GetAll(ctx, q, &players)
	if err != nil {
		return ret, err
	}

	if len(players) == 0 {
		return ret, datastore.ErrNoSuchEntity
	}

	return players, nil
}

func GetPlayers(order string, limit int) (players []Player, err error) {

	client, ctx, err := getClient()
	if err != nil {
		return players, err
	}

	q := datastore.NewQuery(KindPlayer).Order(order)

	if limit > 0 {
		q = q.Limit(limit)
	}

	_, err = client.GetAll(ctx, q, &players)

	return players, err
}

func GetPlayersByIDs(ids []int64) (friends []Player, err error) {

	if len(ids) > 1000 {
		return friends, ErrorTooMany
	}

	client, ctx, err := getClient()
	if err != nil {
		return friends, err
	}

	var keys []*datastore.Key
	for _, v := range ids {
		key := datastore.NameKey(KindPlayer, strconv.FormatInt(v, 10), nil)
		keys = append(keys, key)
	}

	friends = make([]Player, len(keys))
	err = client.GetMulti(ctx, keys, friends)
	if err != nil && !strings.Contains(err.Error(), "no such entity") {
		return friends, err
	}

	return friends, nil
}

func CountPlayers() (count int, err error) {

	if cachePlayersCount == 0 {

		client, ctx, err := getClient()
		if err != nil {
			return count, err
		}

		q := datastore.NewQuery(KindPlayer)
		cachePlayersCount, err = client.Count(ctx, q)
		if err != nil {
			return count, err
		}
	}

	return cachePlayersCount, nil
}

func (p *Player) Update(userAgent string) (errs []error) {

	if !IsValidPlayerID(p.PlayerID) {
		return []error{ErrInvalidID}
	}

	if helpers.IsBot(userAgent) {
		return []error{}
	}

	if p.UpdatedAt.Unix() > (time.Now().Unix() - int64(60*60*24)) { // 1 Day
		return []error{}
	}

	var err error
	var wg sync.WaitGroup

	// Get summary
	wg.Add(1)
	go func(p *Player) {
		err = p.updateSummary()
		logger.Error(err)
		wg.Done()
	}(p)

	// Get games
	wg.Add(1)
	go func(p *Player) {
		err = p.updateGames()
		logger.Error(err)
		wg.Done()
	}(p)

	// Get recent games
	wg.Add(1)
	go func(p *Player) {
		err = p.updateRecentGames()
		logger.Error(err)
		wg.Done()
	}(p)

	// Get badges
	wg.Add(1)
	go func(p *Player) {
		err = p.updateBades()
		logger.Error(err)
		wg.Done()
	}(p)

	// Get friends
	wg.Add(1)
	go func(p *Player) {
		err = p.updateFriends()
		logger.Error(err)
		wg.Done()
	}(p)

	// Get level
	wg.Add(1)
	go func(p *Player) {
		err = p.updateLevel()
		logger.Error(err)
		wg.Done()
	}(p)

	// Get bans
	wg.Add(1)
	go func(p *Player) {
		err = p.updateBans()
		logger.Error(err)
		wg.Done()
	}(p)

	// Get groups
	wg.Add(1)
	go func(p *Player) {
		p.updateGroups()
		logger.Error(err)
		wg.Done()
	}(p)

	// Wait
	wg.Wait()

	err = p.Save()
	if err != nil {
		errs = append(errs, err)
	}

	return errs
}

func (p *Player) updateSummary() (error) {

	summary, _, err := steami.Steam().GetPlayer(p.PlayerID)
	if err != nil {
		return err
	}

	p.Avatar = strings.Replace(summary.AvatarFull, "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/", "", 1)
	p.VanintyURL = path.Base(summary.ProfileURL)
	p.RealName = summary.RealName
	p.CountryCode = summary.LOCCountryCode
	p.StateCode = summary.LOCStateCode
	p.PersonaName = summary.PersonaName
	p.TimeCreated = time.Unix(summary.TimeCreated, 0)
	p.LastLogOff = time.Unix(summary.LastLogOff, 0)
	p.PrimaryClanID = summary.PrimaryClanID

	return err
}

func (p *Player) updateGames() (error) {

	resp, _, err := steami.Steam().GetOwnedGames(p.PlayerID)
	if err != nil {
		return err
	}

	p.GamesCount = len(resp.Games)

	// Get playtime
	var playtime = 0
	for _, v := range resp.Games {
		playtime = playtime + v.PlaytimeForever
	}

	p.PlayTime = playtime

	// Encode to JSON bytes
	bytes, err := json.Marshal(resp.Games)
	if err != nil {
		return err
	}

	// Upload
	if len(bytes) > maxBytesToStore {
		storagePath := storage.PathGames(p.PlayerID)
		err = storage.Upload(storagePath, bytes, false)
		if err != nil {
			return err
		}
		p.Games = storagePath
	} else {
		p.Games = string(bytes)
	}

	return nil
}

func (p *Player) updateRecentGames() (error) {

	recentResponse, _, err := steami.Steam().GetRecentlyPlayedGames(p.PlayerID)
	if err != nil {
		return err
	}

	// Encode to JSON bytes
	bytes, err := json.Marshal(recentResponse.Games)
	if err != nil {
		return err
	}

	// Upload
	if len(bytes) > maxBytesToStore {
		storagePath := storage.PathRecentGames(p.PlayerID)
		err = storage.Upload(storagePath, bytes, false)
		if err != nil {
			return err
		}
		p.GamesRecent = storagePath
	} else {
		p.GamesRecent = string(bytes)
	}

	return nil
}

func (p *Player) updateBades() (error) {

	badgesResponse, _, err := steami.Steam().GetBadges(p.PlayerID)
	if err != nil {
		return err
	}

	p.BadgesCount = len(badgesResponse.Badges)

	// Encode to JSON bytes
	bytes, err := json.Marshal(badgesResponse)
	if err != nil {
		return err
	}

	// Upload
	if len(bytes) > maxBytesToStore {
		storagePath := storage.PathBadges(p.PlayerID)
		err = storage.Upload(storagePath, bytes, false)
		if err != nil {
			return err
		}
		p.Badges = storagePath
	} else {
		p.Badges = string(bytes)
	}

	return nil
}

func (p *Player) updateFriends() (error) {

	resp, _, err := steami.Steam().GetFriendList(p.PlayerID)
	if err != nil {
		return err
	}

	p.FriendsCount = len(resp.Friends)

	// Encode to JSON bytes
	bytes, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	// Upload
	if len(bytes) > maxBytesToStore {
		storagePath := storage.PathFriends(p.PlayerID)
		err = storage.Upload(storagePath, bytes, false)
		if err != nil {
			return err
		}
		p.Friends = storagePath
	} else {
		p.Friends = string(bytes)
	}

	return nil
}

func (p *Player) updateLevel() (error) {

	level, _, err := steami.Steam().GetSteamLevel(p.PlayerID)
	if err != nil {
		return err
	}

	p.Level = level

	return nil
}

func (p *Player) updateBans() (error) {

	bans, _, err := steami.Steam().GetPlayerBans(p.PlayerID)
	if err != nil {
		return err
	}

	p.NumberOfGameBans = bans.NumberOfGameBans
	p.NumberOfVACBans = bans.NumberOfVACBans

	// Encode to JSON bytes
	bytes, err := json.Marshal(bans)
	if err != nil {
		return err
	}

	p.Bans = string(bytes)

	return nil
}

func (p *Player) updateGroups() (error) {

	resp, _, err := steami.Steam().GetUserGroupList(p.PlayerID)
	if err != nil {
		return err
	}

	p.Groups = resp.GetIDs()

	return nil
}

func (p *Player) Save() (err error) {

	if !IsValidPlayerID(p.PlayerID) {
		return ErrInvalidID
	}

	// Fix dates
	p.UpdatedAt = time.Now()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now()
	}

	_, err = SaveKind(p.GetKey(), p)
	if err != nil {
		return err
	}

	return nil
}
