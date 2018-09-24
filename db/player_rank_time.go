package db

import (
	"strconv"
	"time"

	"cloud.google.com/go/datastore"
)

type PlayerRankTime struct {
	PlayerID     int64     `datastore:"player_id"`
	Time         time.Time `datastore:"player_id"`
	LevelRank    int       `datastore:"level_rank,noindex"`
	GamesRank    int       `datastore:"games_rank,noindex"`
	BadgesRank   int       `datastore:"badges_rank,noindex"`
	PlayTimeRank int       `datastore:"play_time_rank,noindex"`
	FriendsRank  int       `datastore:"friends_rank,noindex"`

	Rank string `datastore:"-"`
}

func (p PlayerRankTime) GetKey() (key *datastore.Key) {
	return datastore.NameKey(KindPlayerRankTime, strconv.FormatInt(p.PlayerID, 10), nil)
}
