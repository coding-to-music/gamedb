package queue

import (
	"strconv"
	"strings"

	"github.com/Jleagle/rabbit-go"
	roman "github.com/StefanSchroeder/Golang-Roman"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.uber.org/zap"
)

type AppsSearchMessage struct {
	App   *mongo.App `json:"app"`
	AppID int        `json:"app_id"`
}

func (m AppsSearchMessage) Queue() rabbit.QueueName {
	return QueueAppsSearch
}

func appsSearchHandler(message *rabbit.Message) {

	payload := AppsSearchMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		zap.L().Error(err.Error(), zap.ByteString("message", message.Message.Body))
		sendToFailQueue(message)
		return
	}

	var mongoApp mongo.App

	if payload.AppID > 0 {

		mongoApp, err = mongo.GetApp(payload.AppID)
		if err != nil {
			zap.L().Error(err.Error(), zap.ByteString("message", message.Message.Body))
			sendToRetryQueue(message)
			return
		}

	} else if payload.App != nil {

		mongoApp = *payload.App

	} else {

		zap.S().Error(message.Message.Body)
		sendToFailQueue(message)
		return
	}

	app := elasticsearch.App{}
	app.AchievementsAvg = mongoApp.AchievementsAverageCompletion
	app.AchievementsCount = mongoApp.AchievementsCount
	app.AchievementsIcons = mongoApp.Achievements
	app.Aliases = makeAppAliases(mongoApp.ID, mongoApp.Name)
	app.Categories = mongoApp.Categories
	app.Developers = mongoApp.Developers
	app.FollowersCount = mongoApp.GroupFollowers
	app.Genres = mongoApp.Genres
	app.Icon = mongoApp.Icon
	app.ID = mongoApp.ID
	app.Name = mongoApp.Name
	app.Platforms = mongoApp.Platforms
	app.PlayersCount = mongoApp.PlayerPeakWeek
	app.Prices = mongoApp.Prices
	app.Publishers = mongoApp.Publishers
	app.ReleaseDate = mongoApp.ReleaseDateUnix
	app.ReviewScore = mongoApp.ReviewsScore
	app.Tags = mongoApp.Tags
	app.Trend = mongoApp.PlayerTrend
	app.Type = mongoApp.Type
	app.WishlistAvg = mongoApp.WishlistAvgPosition
	app.WishlistCount = mongoApp.WishlistCount

	err = elasticsearch.IndexApp(app)
	if err != nil {
		zap.S().Error(err)
		sendToRetryQueue(message)
		return
	}

	message.Ack()
}

var aliasMap = map[int][]string{
	813780:  {"aoe", "aoe2"},   // Age of Empires II: Definitive Edition
	221380:  {"aoe", "aoe2"},   // Age of Empires II (2013)
	105450:  {"aoe", "aoe3"},   // Age of Empires® III: Complete Collection
	1017900: {"aoe"},           // Age of Empires: Definitive Edition
	105430:  {"aoe"},           // Age of Empires Online
	346110:  {"ark"},           // ARK: Survival Evolved
	49520:   {"bl", "bl2"},     // Borderlands 2
	397540:  {"bl", "bl3"},     // Borderlands 3
	8980:    {"bl"},            // Borderlands GOTY
	570:     {"dota", "dota2"}, // Dota 2
	8500:    {"eve"},           // EVE Online
	24240:   {"pd", "pdth"},    // PAYDAY: The Heist
	218620:  {"pd", "pd2"},     // PAYDAY 2
	578080:  {"pubg"},          // PLAYERUNKNOWN'S BATTLEGROUNDS
	3900:    {"civ", "civ4"},   // Sid Meier's Civilization IV
	8930:    {"civ", "civ5"},   // Sid Meier's Civilization V
	289070:  {"civ", "civ6"},   // Sid Meier's Civilization VI
	359550:  {"r6"},            // Tom Clancy's Rainbow Six Siege
}

func makeAppAliases(ID int, name string) (aliases []string) {

	if val, ok := aliasMap[ID]; ok {
		aliases = val
	}

	for _, convertRoman := range []bool{true, false} {
		for _, trimSmall := range []bool{true, false} {

			var alias []string
			for _, v := range helpers.RegexNonAlphaNumeric.Split(name, -1) {

				if v == "" {
					continue
				}

				if convertRoman && helpers.RegexSmallRomanOnly.MatchString(v) {
					v = strconv.Itoa(roman.Arabic(v))
				}

				if helpers.RegexIntsOnly.MatchString(v) {
					alias = append(alias, v)
				} else if !trimSmall || len(v) > 1 {
					alias = append(alias, strings.ToLower(v[0:1]))
				}
			}
			if len(alias) > 1 {
				aliases = append(aliases, strings.Join(alias, ""))
			}
		}
	}

	return helpers.UniqueString(aliases)
}
