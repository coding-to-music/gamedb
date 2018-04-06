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
	"github.com/gosimple/slug"
	"github.com/steam-authority/steam-authority/helpers"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/steam"
	"github.com/steam-authority/steam-authority/storage"
)

type Player struct {
	CreatedAt        time.Time                   `datastore:"created_at"`               //
	UpdatedAt        time.Time                   `datastore:"updated_at"`               //
	FriendsAddedAt   time.Time                   `datastore:"friends_added_at,noindex"` //
	PlayerID         int                         `datastore:"player_id"`                //
	VanintyURL       string                      `datastore:"vanity_url"`               //
	Avatar           string                      `datastore:"avatar,noindex"`           //
	PersonaName      string                      `datastore:"persona_name,noindex"`     //
	RealName         string                      `datastore:"real_name,noindex"`        //
	CountryCode      string                      `datastore:"country_code"`             //
	StateCode        string                      `datastore:"status_code"`              //
	Level            int                         `datastore:"level"`                    //
	Games            string                      `datastore:"games,noindex"`            // JSON
	GamesRecent      []steam.RecentlyPlayedGame  `datastore:"games_recent,noindex"`     //
	GamesCount       int                         `datastore:"games_count"`              //
	Badges           steam.BadgesResponse        `datastore:"badges,noindex"`           //
	BadgesCount      int                         `datastore:"badges_count"`             //
	PlayTime         int                         `datastore:"play_time"`                //
	TimeCreated      time.Time                   `datastore:"time_created"`             //
	LastLogOff       time.Time                   `datastore:"time_logged_off,noindex"`  //
	PrimaryClanID    int                         `datastore:"primary_clan_id,noindex"`  //
	Friends          []steam.GetFriendListFriend `datastore:"friends,noindex"`          //
	FriendsCount     int                         `datastore:"friends_count"`            //
	Donated          int                         `datastore:"donated"`                  //
	Bans             steam.GetPlayerBanResponse  `datastore:"bans"`                     //
	NumberOfVACBans  int                         `datastore:"bans_cav"`                 //
	NumberOfGameBans int                         `datastore:"bans_game"`                //
	Groups           []int                       `datastore:"groups,noindex"`           //
}

func (p Player) GetKey() (key *datastore.Key) {
	return datastore.NameKey(KindPlayer, strconv.Itoa(p.PlayerID), nil)
}

func (p Player) GetPath() string {
	return "/players/" + strconv.Itoa(p.PlayerID) + "/" + slug.Make(p.PersonaName)
}

func (p Player) GetSteamTimeUnix() (int64) {
	return p.TimeCreated.Unix()
}

func (p Player) GetSteamTimeNice() (string) {
	return p.TimeCreated.Format(time.RFC822)
}

func (p Player) GetLogoffUnix() (int64) {
	return p.LastLogOff.Unix()
}

func (p Player) GetLogoffNice() (string) {
	return p.LastLogOff.Format(time.RFC822)
}

func (p Player) GetUpdatedUnix() (int64) {
	return p.UpdatedAt.Unix()
}

func (p Player) GetUpdatedNice() (string) {
	return p.UpdatedAt.Format(time.RFC822)
}

func (p Player) GetSteamCommunityLink() string {
	return "http://steamcommunity.com/profiles/" + strconv.Itoa(p.PlayerID)
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

func (p Player) GetGames() (x []steam.OwnedGame) {

	var bytes []byte
	var err error

	if strings.HasPrefix(p.Games, "/") {
		bytes, err = storage.DownloadGamesJson(p.PlayerID)
		if err != nil {
			logger.Error(err)
		}
	} else {
		bytes = []byte(p.Games)
	}

	err = json.Unmarshal(bytes, &x)
	if err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			logger.Info(err.Error() + " - " + string(bytes))
		} else {
			logger.Error(err)
		}
	}

	return x
}

func (p Player) shouldUpdate() bool {

	return true

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

	client, ctx, err := getDSClient()
	if err != nil {
		return ret, err
	}

	key := datastore.NameKey(KindPlayer, strconv.Itoa(id), nil)

	player := new(Player)
	player.PlayerID = id

	err = client.Get(ctx, key, player)
	if err != nil {

		if err.Error() == ErrorNotFound {
			return player, nil
		}
		return player, err
	}

	return player, nil
}

func GetPlayerByName(name string) (ret Player, err error) {

	client, ctx, err := getDSClient()
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
		return players[0], nil
	}

	return ret, errors.New(ErrorNotFound)

}

func GetPlayers(order string, limit int) (players []Player, err error) {

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
		return friends, errors.New("too many players")
	}

	client, ctx, err := getDSClient()
	if err != nil {
		return friends, err
	}

	var keys []*datastore.Key
	for _, v := range ids {
		key := datastore.NameKey(KindPlayer, strconv.Itoa(v), nil)
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

func (p *Player) UpdateIfNeeded() (errs []error) {

	if p.shouldUpdate() {

		var wg sync.WaitGroup
		var errs []error
		var err error

		// Get summary
		wg.Add(1)
		go func(p *Player) {

			summary, err := steam.GetPlayerSummaries(p.PlayerID)
			if err != nil {
				if err.Error() == steam.ErrInvalidJson {
					errs = append(errs, err)
				} else if !strings.HasPrefix(err.Error(), "not found in steam") {
					logger.Error(err)
				}
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

			wg.Done()
		}(p)

		// Get games
		wg.Add(1)
		go func(p *Player) {

			gamesResponse, err := steam.GetOwnedGames(p.PlayerID)
			if err != nil {
				if err.Error() == steam.ErrInvalidJson {
					errs = append(errs, err)
				} else {
					logger.Error(err)
				}
			}

			p.GamesCount = len(gamesResponse)

			// Get playtime
			var playtime = 0
			for _, v := range gamesResponse {
				playtime = playtime + v.PlaytimeForever
			}

			p.PlayTime = playtime

			// Encode to JSON bytes
			bytes, err := json.Marshal(gamesResponse)
			if err != nil {
				logger.Error(err)
			}

			if len(bytes) > 1024*10 {
				p.Games = storage.UploadGamesJson(p.PlayerID, bytes)
			} else {
				p.Games = string(bytes)
			}

			wg.Done()
		}(p)

		// Get recent games
		wg.Add(1)
		go func(p *Player) {

			recentGames, err := steam.GetRecentlyPlayedGames(p.PlayerID)
			if err != nil {
				if err.Error() == steam.ErrInvalidJson {
					errs = append(errs, err)
				} else {
					logger.Error(err)
				}
			}

			p.GamesRecent = recentGames

			wg.Done()
		}(p)

		// Get badges
		wg.Add(1)
		go func(p *Player) {

			badges, err := steam.GetBadges(p.PlayerID)
			if err != nil {
				if err.Error() == steam.ErrInvalidJson {
					errs = append(errs, err)
				} else {
					logger.Error(err)
				}
			}

			p.Badges = badges
			p.BadgesCount = len(badges.Badges)

			wg.Done()
		}(p)

		// Get friends
		wg.Add(1)
		go func(p *Player) {

			friends, err := steam.GetFriendList(p.PlayerID)
			if err != nil {
				if err.Error() == steam.ErrInvalidJson {
					errs = append(errs, err)
				} else {
					logger.Error(err)
				}
			}

			p.Friends = friends
			p.FriendsCount = len(friends)

			wg.Done()
		}(p)

		// Get level
		wg.Add(1)
		go func(p *Player) {

			level, err := steam.GetSteamLevel(p.PlayerID)
			if err != nil {
				if err.Error() == steam.ErrInvalidJson {
					errs = append(errs, err)
				} else {
					logger.Error(err)
				}
			}

			p.Level = level

			wg.Done()
		}(p)

		// Get bans
		wg.Add(1)
		go func(p *Player) {

			bans, err := steam.GetPlayerBans(p.PlayerID)
			if err != nil {
				if err.Error() == steam.ErrInvalidJson {
					errs = append(errs, err)
				} else {
					logger.Error(err)
				}
			}

			p.Bans = bans
			p.NumberOfGameBans = bans.NumberOfGameBans
			p.NumberOfVACBans = bans.NumberOfVACBans

			wg.Done()
		}(p)

		// Get groups
		wg.Add(1)
		go func(p *Player) {

			groups, err := steam.GetUserGroupList(p.PlayerID)
			if err != nil {
				if err.Error() == steam.ErrInvalidJson {
					errs = append(errs, err)
				} else {
					logger.Error(err)
				}
			}

			p.Groups = groups

			wg.Done()
		}(p)

		wg.Wait()

		// Fix dates
		p.UpdatedAt = time.Now()
		if p.CreatedAt.IsZero() {
			p.CreatedAt = time.Now()
		}

		err = p.Save()
		if err != nil {
			errs = append(errs, err)
		}
	}

	return
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
