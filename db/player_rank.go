package db

import (
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/steam-authority/steam-authority/helpers"
	"github.com/steam-authority/steam-authority/memcache"
)

const PlayersToRank = 1000

type PlayerRank struct {
	CreatedAt   time.Time `datastore:"created_at,noindex"`
	UpdatedAt   time.Time `datastore:"updated_at,noindex"`
	PlayerID    int64     `datastore:"player_id,noindex"`
	VanintyURL  string    `datastore:"vality_url,noindex"`
	Avatar      string    `datastore:"avatar,noindex"`
	PersonaName string    `datastore:"persona_name,noindex"`
	CountryCode string    `datastore:"country_code"`

	// Ranks
	Level        int `datastore:"level"`
	LevelRank    int `datastore:"level_rank"`
	Games        int `datastore:"games"`
	GamesRank    int `datastore:"games_rank"`
	Badges       int `datastore:"badges"`
	BadgesRank   int `datastore:"badges_rank"`
	PlayTime     int `datastore:"play_time"`
	PlayTimeRank int `datastore:"play_time_rank"`
	Friends      int `datastore:"friends"`
	FriendsRank  int `datastore:"friends_rank"`
}

func (rank PlayerRank) GetKey() (key *datastore.Key) {
	return datastore.NameKey(KindPlayerRank, strconv.FormatInt(rank.PlayerID, 10), nil)
}

func (rank PlayerRank) GetAvatar() string {
	if strings.HasPrefix(rank.Avatar, "http") {
		return rank.Avatar
	} else if rank.Avatar != "" {
		return "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/" + rank.Avatar
	} else {
		return rank.GetDefaultAvatar()
	}
}

func (rank PlayerRank) GetAvatar2() string {
	return helpers.GetAvatar2(rank.Level)
}

func (rank PlayerRank) GetDefaultAvatar() string {
	return "/assets/img/no-player-image.jpg"
}

func (rank PlayerRank) GetFlag() string {

	if rank.CountryCode == "" {
		return ""
	}

	return "/assets/img/flags/" + strings.ToLower(rank.CountryCode) + ".png"
}

func (rank PlayerRank) GetCountry() string {
	return helpers.CountryCodeToName(rank.CountryCode)
}

func (rank PlayerRank) GetTimeShort() (ret string) {
	return helpers.GetTimeShort(rank.PlayTime, 2)
}

func (rank PlayerRank) GetTimeLong() (ret string) {
	return helpers.GetTimeLong(rank.PlayTime, 5)
}

func GetRank(playerID int64) (rank *PlayerRank, err error) {

	client, context, err := GetDSClient()
	if err != nil {
		return rank, err
	}

	key := datastore.NameKey(KindPlayerRank, strconv.FormatInt(playerID, 10), nil)

	rank = new(PlayerRank)
	rank.PlayerID = playerID

	err = client.Get(context, key, rank)
	return rank, err
}

// Returns as much data as it can if there is an error
func GetRankKeys() (keysMap map[int64]*datastore.Key, err error) {

	keysMap = make(map[int64]*datastore.Key)

	client, ctx, err := GetDSClient()
	if err != nil {
		return keysMap, err
	}

	q := datastore.NewQuery(KindPlayerRank).KeysOnly()
	keys, err := client.GetAll(ctx, q, nil)
	if err != nil {
		return
	}

	var errors []error
	for _, v := range keys {
		playerId, err := strconv.ParseInt(v.Name, 10, 64)
		if err != nil {
			errors = append(errors, err)
		} else {
			keysMap[playerId] = v
		}
	}

	if len(errors) > 0 {
		return keysMap, errors[0]
	}

	return keysMap, nil
}

func CountRanks() (count int, err error) {

	return PlayersToRank, nil

	return memcache.GetSetInt(memcache.RanksCount, func() (count int, err error) {

		client, ctx, err := GetDSClient()
		if err != nil {
			return count, err
		}

		q := datastore.NewQuery(KindPlayerRank)
		count, err = client.Count(ctx, q)
		return count, err
	})
}

func NewRankFromPlayer(player Player) (rank *PlayerRank) {

	rank = new(PlayerRank)

	// Profile
	rank.CreatedAt = time.Now()
	rank.UpdatedAt = time.Now()
	rank.PlayerID = player.PlayerID
	rank.VanintyURL = player.VanintyURL
	rank.Avatar = player.Avatar
	rank.PersonaName = player.PersonaName
	rank.CountryCode = player.CountryCode

	// Rankable
	rank.Level = player.Level
	rank.Games = player.GamesCount
	rank.Badges = player.BadgesCount
	rank.PlayTime = player.PlayTime
	rank.Friends = player.FriendsCount

	return rank
}
