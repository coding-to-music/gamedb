package mongo

import (
	"strings"
	"time"

	"github.com/gamedb/website/helpers"
)

const PlayersToRank = 1000

type PlayerRank struct {
	CreatedAt   time.Time `bson:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at"`
	PlayerID    int64     `bson:"player_id"`
	VanintyURL  string    `bson:"vality_url"`
	Avatar      string    `bson:"avatar"`
	PersonaName string    `bson:"persona_name"`
	CountryCode string    `bson:"country_code"`

	// Ranks
	Level        int `bson:"level"`
	LevelRank    int `bson:"level_rank"`
	Games        int `bson:"games"`
	GamesRank    int `bson:"games_rank"`
	Badges       int `bson:"badges"`
	BadgesRank   int `bson:"badges_rank"`
	PlayTime     int `bson:"play_time"`
	PlayTimeRank int `bson:"play_time_rank"`
	Friends      int `bson:"friends"`
	FriendsRank  int `bson:"friends_rank"`
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

// func GetRank(playerID int64) (rank PlayerRank, err error) {
//
// 	client, context, err := GetDSClient()
// 	if err != nil {
// 		return rank, err
// 	}
//
// 	key := datastore.NameKey(KindPlayerRank, strconv.FormatInt(playerID, 10), nil)
//
// 	rank = PlayerRank{}
// 	rank.PlayerID = playerID
//
// 	err = client.Get(context, key, &rank)
// 	return rank, err
// }
//
// // Returns as much data as it can if there is an error
// func GetRankKeys() (keysMap map[int64]*datastore.Key, err error) {
//
// 	keysMap = make(map[int64]*datastore.Key)
//
// 	client, ctx, err := GetDSClient()
// 	if err != nil {
// 		return keysMap, err
// 	}
//
// 	q := datastore.NewQuery(KindPlayerRank).KeysOnly()
// 	keys, err := client.GetAll(ctx, q, nil)
// 	if err != nil {
// 		return
// 	}
//
// 	var errors []error
// 	for _, v := range keys {
// 		playerID, err := strconv.ParseInt(v.Name, 10, 64)
// 		if err != nil {
// 			errors = append(errors, err)
// 		} else {
// 			keysMap[playerID] = v
// 		}
// 	}
//
// 	if len(errors) > 0 {
// 		return keysMap, errors[0]
// 	}
//
// 	return keysMap, nil
// }

func CountRanks() (count int64, err error) {

	var item = helpers.MemcacheRanksCount

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &count, func() (interface{}, error) {

		return CountDocuments(CollectionPlayerRanks, nil)
	})

	return count, err
}

func NewRankFromPlayer(player Player) (rank *PlayerRank) {

	rank = new(PlayerRank)

	// Profile
	rank.CreatedAt = time.Now()
	rank.UpdatedAt = time.Now()
	rank.PlayerID = player.ID
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
