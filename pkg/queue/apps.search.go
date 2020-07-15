package queue

import (
	"strconv"
	"strings"

	"github.com/Jleagle/rabbit-go"
	roman "github.com/StefanSchroeder/Golang-Roman"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
)

type AppsSearchMessage struct {
	App mongo.App `json:"app"`
}

func (m AppsSearchMessage) Queue() rabbit.QueueName {
	return QueueAppsSearch
}

func appsSearchHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		payload := AppsSearchMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			continue
		}

		app := elasticsearch.App{}
		app.ID = payload.App.ID
		app.Name = payload.App.Name
		app.Players = payload.App.PlayerPeakWeek
		app.Icon = payload.App.Icon
		app.Followers = payload.App.GroupFollowers
		app.ReviewScore = payload.App.ReviewsScore
		app.Prices = payload.App.Prices
		app.Tags = payload.App.Tags
		app.Genres = payload.App.Genres
		app.Categories = payload.App.Categories
		app.Publishers = payload.App.Publishers
		app.Developers = payload.App.Developers
		app.Type = payload.App.Type
		app.Platforms = payload.App.Platforms
		app.Aliases = makeAppAliases(app)

		err = elasticsearch.IndexApp(app)
		if err != nil {
			log.Err(err)
			sendToRetryQueue(message)
			continue
		}

		message.Ack(false)
	}
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

func makeAppAliases(app elasticsearch.App) (aliases []string) {

	if val, ok := aliasMap[app.ID]; ok {
		aliases = val
	}

	for _, convertRoman := range []bool{true, false} {
		for _, trimSmall := range []bool{true, false} {

			var alias []string
			for _, v := range helpers.RegexNonAlphaNumeric.Split(app.Name, -1) {

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
