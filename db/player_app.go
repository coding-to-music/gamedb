package db

import (
	"strconv"

	"cloud.google.com/go/datastore"
	"github.com/steam-authority/steam-authority/helpers"
)

type PlayerApp struct {
	PlayerID     int64   `datastore:"player_id"`
	AppID        int     `datastore:"app_id,noindex"`
	AppName      string  `datastore:"app_name"`
	AppIcon      string  `datastore:"app_icon,noindex"`
	AppTime      int     `datastore:"app_time"`
	AppPrice     int     `datastore:"app_price"`
	AppPriceHour float64 `datastore:"app_price_hour"`
}

func (p PlayerApp) GetKey() (key *datastore.Key) {
	return datastore.NameKey(KindPlayerApp, strconv.FormatInt(p.PlayerID, 10)+"-"+strconv.Itoa(p.AppID), nil)
}

func (p PlayerApp) GetIcon() string {

	if p.AppIcon == "" {
		return "/assets/img/no-player-image.jpg" // todo, fix to right image
	}

	return "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/" + strconv.Itoa(p.AppID) + "/" + p.AppIcon + ".jpg"
}

func (p PlayerApp) GetTimeNice() string {

	return helpers.GetTimeShort(p.AppTime, 2)
}

func (p PlayerApp) GetPriceHourFormatted() string {

	x := p.AppPriceHour
	if x == -1 {
		return "∞"
	}

	return strconv.FormatFloat(helpers.DollarsFloat(x), 'f', 2, 64)
}

func (p PlayerApp) GetPriceHourSort() string {

	x := p.AppPriceHour
	if x == -1 {
		return "1000000"
	}

	return strconv.FormatFloat(helpers.DollarsFloat(x), 'f', 2, 64)
}

func (p *PlayerApp) SetPriceHour() {

	if p.AppPrice == 0 {
		return
	}

	if p.AppTime == 0 {
		return
	}

	p.AppPriceHour = float64(p.AppPrice) / (float64(p.AppTime) / 60)
}

func GetPlayerApps(playerID int64, sort string, limit int) (apps []PlayerApp, err error) {

	client, ctx, err := GetDSClient()
	if err != nil {
		return apps, err
	}

	q := datastore.NewQuery(KindPlayerApp).Filter("player_id =", playerID).Order(sort)

	if limit > 0 {
		q.Limit(limit)
	}

	_, err = client.GetAll(ctx, q, &apps)
	if err != nil {
		return
	}

	return apps, err
}

func BulkSavePlayerApps(apps []*PlayerApp) (err error) {

	if len(apps) == 0 {
		return nil
	}

	client, ctx, err := GetDSClient()
	if err != nil {
		return err
	}

	chunks := chunkPlayerApps(apps, 500)

	for _, chunk := range chunks {

		keys := make([]*datastore.Key, 0, len(chunk))
		for _, v := range chunk {
			keys = append(keys, v.GetKey())
		}

		_, err = client.PutMulti(ctx, keys, chunk)
		if err != nil {
			return err
		}
	}

	return nil
}

func chunkPlayerApps(changes []*PlayerApp, chunkSize int) (divided [][]*PlayerApp) {

	for i := 0; i < len(changes); i += chunkSize {
		end := i + chunkSize

		if end > len(changes) {
			end = len(changes)
		}

		divided = append(divided, changes[i:end])
	}

	return divided
}
