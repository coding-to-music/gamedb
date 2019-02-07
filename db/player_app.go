package db

import (
	"reflect"
	"strconv"
	"strings"

	"cloud.google.com/go/datastore"
	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
)

type PlayerApp struct {
	PlayerID     int64                  `datastore:"player_id"`
	AppID        int                    `datastore:"app_id,noindex"`
	AppName      string                 `datastore:"app_name"`
	AppIcon      string                 `datastore:"app_icon,noindex"`
	AppTime      int                    `datastore:"app_time"`
	AppPrices    CountryCodeIntStruct   `datastore:"app_price,flatten"`
	AppPriceHour CountryCodeFloatStruct `datastore:"app_price_hour,flatten"`
}

func (p PlayerApp) GetKey() (key *datastore.Key) {
	return datastore.NameKey(KindPlayerApp, strconv.FormatInt(p.PlayerID, 10)+"-"+strconv.Itoa(p.AppID), nil)
}

func (p PlayerApp) GetPath() string {
	return GetAppPath(p.AppID, p.AppName)
}

func (p PlayerApp) GetIcon() string {

	if p.AppIcon == "" {
		return "/assets/img/no-player-image.jpg"
	}
	return "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/" + strconv.Itoa(p.AppID) + "/" + p.AppIcon + ".jpg"
}

func (p PlayerApp) GetTimeNice() string {

	return helpers.GetTimeShort(p.AppTime, 2)
}

func (p PlayerApp) GetPriceFormatted(code steam.CountryCode) string {

	s := reflect.Indirect(reflect.ValueOf(p.AppPrices))
	f := s.FieldByName(string(code))

	if f.IsNil() {
		return ""
	}

	locale, err := helpers.GetLocaleFromCountry(code)
	log.Err(err)

	return locale.Format(int(f.Elem().Int()))
}

func (p PlayerApp) GetPriceHourFormatted(code steam.CountryCode) string {

	s := reflect.Indirect(reflect.ValueOf(p.AppPriceHour))
	f := s.FieldByName(string(code))

	if f.IsNil() {
		return ""
	}

	locale, err := helpers.GetLocaleFromCountry(code)
	log.Err(err)

	val := f.Elem().Float()
	if val < 0 {
		return "∞"
	}

	return locale.FormatFloat(val)
}

func (p PlayerApp) OutputForJSON(code steam.CountryCode) (output []interface{}) {

	return []interface{}{
		p.AppID,
		p.AppName,
		p.GetIcon(),
		p.AppTime,
		p.GetTimeNice(),
		p.GetPriceFormatted(code),
		p.GetPriceHourFormatted(code),
		p.GetPath(),
	}
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

// These are here because datastore needs a struct to flatten from
type CountryCodeIntStruct struct {
	AE *int
	AR *int
	AU *int
	BR *int
	CA *int
	CH *int
	CL *int
	CN *int
	CO *int
	CR *int
	DE *int
	GB *int
	HK *int
	IL *int
	ID *int
	IN *int
	JP *int
	KR *int
	KW *int
	KZ *int
	MX *int
	MY *int
	NO *int
	NZ *int
	PE *int
	PH *int
	PL *int
	QA *int
	RU *int
	SA *int
	SG *int
	TH *int
	TR *int
	TW *int
	UA *int
	US *int
	UY *int
	VN *int
	ZA *int
}

type CountryCodeFloatStruct struct {
	AE *float64
	AR *float64
	AU *float64
	BR *float64
	CA *float64
	CH *float64
	CL *float64
	CN *float64
	CO *float64
	CR *float64
	DE *float64
	GB *float64
	HK *float64
	IL *float64
	ID *float64
	IN *float64
	JP *float64
	KR *float64
	KW *float64
	KZ *float64
	MX *float64
	MY *float64
	NO *float64
	NZ *float64
	PE *float64
	PH *float64
	PL *float64
	QA *float64
	RU *float64
	SA *float64
	SG *float64
	TH *float64
	TR *float64
	TW *float64
	UA *float64
	US *float64
	UY *float64
	VN *float64
	ZA *float64
}
