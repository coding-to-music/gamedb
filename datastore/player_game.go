package datastore

import (
	"strconv"

	"cloud.google.com/go/datastore"
)

type PlayerGame struct {
	PlayerID     int64   `datastore:"player_id"`
	AppID        int     `datastore:"app_id,noindex"`
	AppName      string  `datastore:"app_name"`
	AppIcon      string  `datastore:"app_icon,noindex"`
	AppPrice     int     `datastore:"app_price"`
	AppTime      int     `datastore:"app_time"`
	AppPriceHour float64 `datastore:"app_price_hour"`
}

func (p PlayerGame) GetKey() (key *datastore.Key) {
	return datastore.NameKey(KindPlayerApp, strconv.FormatInt(p.PlayerID, 10)+"-"+strconv.Itoa(p.AppID), nil)
}

func GetPlayerGames(playerID int64) (changes []Change, err error) {

	client, ctx, err := getClient()
	if err != nil {
		return changes, err
	}

	offset := (page - 1) * 100

	q := datastore.NewQuery(KindChange).Order("-change_id").Limit(limit).Offset(offset)

	client.GetAll(ctx, q, &changes)

	return changes, err
}
