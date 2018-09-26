package db

import (
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/steam-authority/steam-authority/helpers"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/memcache"
)

type Rank struct {
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

func (rank Rank) GetKey() (key *datastore.Key) {
	return datastore.NameKey(KindRank, strconv.FormatInt(rank.PlayerID, 10), nil)
}

func (rank Rank) GetAvatar() string {
	if strings.HasPrefix(rank.Avatar, "http") {
		return rank.Avatar
	} else if rank.Avatar != "" {
		return "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/" + rank.Avatar
	} else {
		return rank.GetDefaultAvatar()
	}
}

func (rank Rank) GetAvatar2() string {
	return helpers.GetAvatar2(rank.Level)
}

func (rank Rank) GetDefaultAvatar() string {
	return "/assets/img/no-player-image.jpg"
}

func (rank Rank) GetFlag() string {

	if rank.CountryCode == "" {
		return ""
	}

	return "/assets/img/flags/" + strings.ToLower(rank.CountryCode) + ".png"
}

func (rank Rank) GetCountry() string {
	return helpers.CountryCodeToName(rank.CountryCode)
}

func (rank Rank) GetTimeShort() (ret string) {
	return helpers.GetTimeShort(rank.PlayTime, 2)
}

func (rank Rank) GetTimeLong() (ret string) {
	return helpers.GetTimeLong(rank.PlayTime, 5)
}

func (rank *Rank) Tidy() *Rank {

	rank.UpdatedAt = time.Now()
	if rank.CreatedAt.IsZero() {
		rank.CreatedAt = time.Now()
	}

	return rank
}

func GetRank(playerID int64) (rank *Rank, err error) {

	client, context, err := GetDSClient()
	if err != nil {
		return rank, err
	}

	key := datastore.NameKey(KindRank, strconv.FormatInt(playerID, 10), nil)

	rank = new(Rank)
	rank.PlayerID = playerID

	err = client.Get(context, key, rank)

	return rank, err
}

func GetRankKeys() (keysMap map[int64]*datastore.Key, err error) {

	keysMap = make(map[int64]*datastore.Key)

	client, ctx, err := GetDSClient()
	if err != nil {
		return keysMap, err
	}

	q := datastore.NewQuery(KindRank).KeysOnly()
	keys, err := client.GetAll(ctx, q, nil)
	if err != nil {
		return
	}

	for _, v := range keys {
		playerId, _ := strconv.ParseInt(v.Name, 10, 64)
		keysMap[playerId] = v
	}

	return keysMap, nil
}

func CountRanks() (count int, err error) {

	err = memcache.GetSet(memcache.AppsCount, &count, func(count interface{}) (err error) {

		client, ctx, err := GetDSClient()
		if err != nil {
			return err
		}

		q := datastore.NewQuery(KindRank)
		count, err = client.Count(ctx, q)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return count, err
	}

	return count, nil
}

func NewRankFromPlayer(player Player) (rank *Rank) {

	rank = new(Rank)

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

func BulkSaveRanks(ranks []*Rank) (err error) {

	if len(ranks) == 0 {
		return nil
	}

	client, context, err := GetDSClient()
	if err != nil {
		return err
	}

	chunks := chunkRanks(ranks, 500)

	for _, v := range chunks {

		keys := make([]*datastore.Key, 0, len(v))
		for _, vv := range v {
			keys = append(keys, vv.GetKey())
		}

		_, err = client.PutMulti(context, keys, v)
		if err != nil {
			logger.Error(err)
		}
	}

	return nil
}

func BulkDeleteRanks(keys map[int64]*datastore.Key) (err error) {

	if len(keys) == 0 {
		return nil
	}

	// Make map a slice
	var keysToDelete []*datastore.Key
	for _, v := range keys {
		keysToDelete = append(keysToDelete, v)
	}

	client, ctx, err := GetDSClient()
	if err != nil {
		return err
	}

	chunks := chunkRankKeys(keysToDelete, 500)

	for _, v := range chunks {

		err = client.DeleteMulti(ctx, v)
		if err != nil {
			return err
		}
	}

	return nil
}

func chunkRanks(ranks []*Rank, chunkSize int) (divided [][]*Rank) {

	for i := 0; i < len(ranks); i += chunkSize {
		end := i + chunkSize

		if end > len(ranks) {
			end = len(ranks)
		}

		divided = append(divided, ranks[i:end])
	}

	return divided
}

func chunkRankKeys(logs []*datastore.Key, chunkSize int) (divided [][]*datastore.Key) {

	for i := 0; i < len(logs); i += chunkSize {
		end := i + chunkSize

		if end > len(logs) {
			end = len(logs)
		}

		divided = append(divided, logs[i:end])
	}

	return divided
}
