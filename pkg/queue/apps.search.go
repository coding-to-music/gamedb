package queue

import (
	"strconv"

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

		row := elastic.App{}
		row.ID = payload.App.ID
		row.Name = payload.App.Name
		row.Players = payload.App.PlayerPeakWeek
		// row.Icon = payload.App.Icon
		// row.Followers = payload.App.GroupFollowers
		// row.Score = payload.App.Score
		// row.Prices = payload.App.Prices

		if val, ok := aliases[payload.App.ID]; ok {
			row.Aliases = val
		}

		err = elastic.SaveToElastic(elastic.IndexApps, strconv.Itoa(payload.App.ID), row)
		if err != nil {
			log.Err(err)
			sendToRetryQueue(message)
			continue
		}

		message.Ack(false)
	}
}
