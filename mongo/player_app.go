package mongo

import (
	"reflect"
	"strconv"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PlayerApp struct {
	PlayerID     int64              ``
	AppID        int                ``
	AppName      string             ``
	AppIcon      string             ``
	AppTime      int                ``
	AppPrices    map[string]int     ``
	AppPriceHour map[string]float32 ``
}

func (pa PlayerApp) Key() interface{} {
	return strconv.FormatInt(pa.PlayerID, 10) + "-" + strconv.Itoa(pa.AppID)
}

func (pa PlayerApp) BSON() (ret interface{}) {

	var prices = bson.M{}
	for k, v := range pa.AppPrices {
		prices[k] = v
	}

	var pricesHour = bson.M{}
	for k, v := range pa.AppPriceHour {
		pricesHour[k] = v
	}

	return bson.M{
		"_id":             pa.Key(),
		"player_id":       pa.PlayerID,
		"app_id":          pa.AppID,
		"app_name":        pa.AppName,
		"app_icon":        pa.AppIcon,
		"app_time":        pa.AppTime,
		"app_prices":      prices,
		"app_prices_hour": pricesHour,
	}
}

func (pa PlayerApp) GetPath() string {
	return helpers.GetAppPath(pa.AppID, pa.AppName)
}

func (pa PlayerApp) GetIcon() string {

	if pa.AppIcon == "" {
		return "/assets/img/no-player-image.jpg"
	}
	return "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/" + strconv.Itoa(pa.AppID) + "/" + pa.AppIcon + ".jpg"
}

func (pa PlayerApp) GetTimeNice() string {

	return helpers.GetTimeShort(pa.AppTime, 2)
}

func (pa PlayerApp) GetPriceFormatted(code steam.CountryCode) string {

	s := reflect.Indirect(reflect.ValueOf(pa.AppPrices))
	f := s.FieldByName(string(code))

	if f.IsNil() {
		return ""
	}

	locale, err := helpers.GetLocaleFromCountry(code)
	log.Err(err)

	return locale.Format(int(f.Elem().Int()))
}

func (pa PlayerApp) GetPriceHourFormatted(code steam.CountryCode) string {

	s := reflect.Indirect(reflect.ValueOf(pa.AppPriceHour))
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

func (pa PlayerApp) OutputForJSON(code steam.CountryCode) (output []interface{}) {

	return []interface{}{
		pa.AppID,
		pa.AppName,
		pa.GetIcon(),
		pa.AppTime,
		pa.GetTimeNice(),
		pa.GetPriceFormatted(code),
		pa.GetPriceHourFormatted(code),
		pa.GetPath(),
	}
}

func GetPlayerAppsByPlayers(playerIDs []int64) (apps []PlayerApp, err error) {

	playersFilter := bson.A{}
	for _, v := range playerIDs {
		playersFilter = append(playersFilter, v)
	}

	return getPlayerApps(0, 0, bson.M{"$or": playersFilter})
}

func GetPlayerAppsByPlayer(playerID int64, offset int64, limit bool, ops *options.FindOptions) (apps []PlayerApp, err error) {

	return getPlayerApps(offset, 100, bson.M{"player_id": playerID})
}

func getPlayerApps(offset int64, limit int64, filter interface{}) (apps []PlayerApp, err error) {

	client, ctx, err := GetMongo()
	if err != nil {
		return apps, err
	}

	ops := options.Find()
	if offset > 0 {
		ops.SetSkip(offset)
	}
	if limit > 0 {
		ops.SetLimit(limit)
	}

	c := client.Database(MongoDatabase, options.Database()).Collection(CollectionPlayerApps)
	cur, err := c.Find(ctx, filter, ops)
	if err != nil {
		return apps, err
	}

	defer func() {
		err = cur.Close(ctx)
		log.Err(err)
	}()

	for cur.Next(ctx) {

		var app PlayerApp
		err := cur.Decode(&app)
		log.Err(err)
		apps = append(apps, app)
	}

	return apps, cur.Err()
}
