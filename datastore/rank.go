package datastore

import (
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/steam-authority/steam-authority/helpers"
	"github.com/steam-authority/steam-authority/logger"
)

var (
	cacheRanksCount int
)

type Rank struct {
	CreatedAt   time.Time `datastore:"created_at,noindex"`
	UpdatedAt   time.Time `datastore:"updated_at,noindex"`
	PlayerID    int       `datastore:"player_id,noindex"`
	VanintyURL  string    `datastore:"vality_url,noindex"`
	Avatar      string    `datastore:"avatar,noindex"`
	PersonaName string    `datastore:"persona_name,noindex"`
	CountryCode string    `datastore:"country_code"`

	// Ranks
	Level        int `datastore:"level"`
	LevelRank    int `datastore:"level_rank"`
	GamesCount   int `datastore:"games"`
	GamesRank    int `datastore:"games_rank"`
	BadgesCount  int `datastore:"badges"`
	BadgesRank   int `datastore:"badges_rank"`
	PlayTime     int `datastore:"play_time"`
	PlayTimeRank int `datastore:"play_time_rank"`
	FriendsCount int `datastore:"friends"`
	FriendsRank  int `datastore:"friends_rank"`

	Rank string `datastore:"-"` // Internal
}

func (rank Rank) GetKey() (key *datastore.Key) {
	return datastore.NameKey(KindRank, strconv.Itoa(rank.PlayerID), nil)
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

func (rank Rank) GetDefaultAvatar() string {
	return "/assets/img/no-player-image.jpg"
}

func (rank Rank) GetFlag() string {
	return "/assets/img/flags/" + strings.ToLower(rank.CountryCode) + ".png"
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

func GetRank(playerID int) (rank *Rank, err error) {

	client, context, err := getClient()
	if err != nil {
		return rank, err
	}

	key := datastore.NameKey(KindRank, strconv.Itoa(playerID), nil)

	rank = new(Rank)
	rank.PlayerID = playerID

	err = client.Get(context, key, rank)

	return rank, err
}

func GetRanksBy(order string, limit int, page int) (ranks []Rank, err error) {

	client, ctx, err := getClient()
	if err != nil {
		return ranks, err
	}

	offset := (page - 1) * limit

	q := datastore.NewQuery(KindRank).Order(order).Limit(limit).Offset(offset)

	client.GetAll(ctx, q, &ranks)

	return ranks, err
}

func GetRankKeys() (keysMap map[int]*datastore.Key, err error) {

	keysMap = make(map[int]*datastore.Key)

	client, ctx, err := getClient()
	if err != nil {
		return keysMap, err
	}

	q := datastore.NewQuery(KindRank).KeysOnly()
	keys, err := client.GetAll(ctx, q, nil)
	if err != nil {
		return keysMap, err
	}

	for _, v := range keys {
		playerId, _ := strconv.Atoi(v.Name)
		keysMap[playerId] = v
	}

	return keysMap, nil
}

func CountRanks() (count int, err error) {

	if cacheRanksCount == 0 {

		client, ctx, err := getClient()
		if err != nil {
			return count, err
		}

		q := datastore.NewQuery(KindRank)
		cacheRanksCount, err = client.Count(ctx, q)
		if err != nil {
			return count, err
		}
	}

	return cacheRanksCount, nil
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
	rank.GamesCount = player.GamesCount
	rank.BadgesCount = player.BadgesCount
	rank.PlayTime = player.PlayTime
	rank.FriendsCount = player.FriendsCount

	return rank
}

func BulkSaveRanks(ranks []*Rank) (err error) {

	ranksLen := len(ranks)
	if ranksLen == 0 {
		return nil
	}

	client, context, err := getClient()
	if err != nil {
		return err
	}

	chunks := chunkRanks(ranks, 500)

	for _, v := range chunks {

		ranksLen := len(v)

		keys := make([]*datastore.Key, 0, ranksLen)
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

func BulkDeleteRanks(keys map[int]*datastore.Key) (err error) {

	// Make map a slice
	var keysToDelete []*datastore.Key
	for _, v := range keys {
		keysToDelete = append(keysToDelete, v)
	}

	client, ctx, err := getClient()
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

func chunkRanks(logs []*Rank, chunkSize int) (divided [][]*Rank) {

	for i := 0; i < len(logs); i += chunkSize {
		end := i + chunkSize

		if end > len(logs) {
			end = len(logs)
		}

		divided = append(divided, logs[i:end])
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
