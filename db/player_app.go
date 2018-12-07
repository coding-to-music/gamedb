package db

import (
	"strconv"
	"strings"

	"cloud.google.com/go/datastore"
	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/helpers"
)

type PlayerApp struct {
	PlayerID     int64                         `datastore:"player_id"`
	AppID        int                           `datastore:"app_id,noindex"`
	AppName      string                        `datastore:"app_name"`
	AppIcon      string                        `datastore:"app_icon,noindex"`
	AppTime      int                           `datastore:"app_time"`
	AppPrices    map[steam.CountryCode]int     `datastore:"app_price"`
	AppPriceHour map[steam.CountryCode]float64 `datastore:"app_price_hour"`
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

func (p PlayerApp) GetPriceHourFormatted(code steam.CountryCode) string {

	priceHour := p.AppPriceHour[code]
	if priceHour == -1 {
		return "∞"
	}
	return helpers.FloatToString(priceHour, 2)
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
