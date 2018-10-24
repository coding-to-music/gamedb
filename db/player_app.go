package db

import (
	"strconv"
	"strings"

	"cloud.google.com/go/datastore"
	"github.com/gamedb/website/helpers"
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

	p.AppPriceHour = (float64(p.AppPrice) / 100) / (float64(p.AppTime) / 60)
}

func ParsePlayerAppKey(key datastore.Key) (playerID int64, appID int, err error) {

	parts := strings.Split(key.Name, "-")

	playerID, err = strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return
	}

	appID, err = strconv.Atoi(parts[1])
	if err != nil {
		return
	}

	return
}
