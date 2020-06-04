package queue

import (
	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/elastic"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
)

type AppsSearchMessage struct {
	App mongo.App `json:"app"`
}

var aliases = map[int][]string{
	10:     {"cs"},
	70:     {"hl"},
	220:    {"hl2"},
	240:    {"css"},
	440:    {"tf2"},
	500:    {"l4d"},
	550:    {"l4d2"},
	570:    {"dota", "dota2"},
	730:    {"csgo", "cs:go"},
	8980:   {"bl"},
	4000:   {"gm"},
	49520:  {"bl2"},
	8500:   {"eve"},
	218620: {"pd2"},
	238960: {"poe"},
	252950: {"rl"},
	271590: {"gta5"},
	346110: {"ark"},
	359550: {"r6"},
	397540: {"bl3"},
	546560: {"hla"},
	985890: {"sor", "sor4"},
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

		app := elastic.App{}
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

		if val, ok := aliases[payload.App.ID]; ok {
			app.Aliases = val
		}

		err = elastic.IndexApp(app)
		if err != nil {
			log.Err(err)
			sendToRetryQueue(message)
			continue
		}

		message.Ack(false)
	}
}
