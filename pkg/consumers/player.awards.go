package consumers

import (
	"strconv"
	"strings"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/influx/schemas"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/steam"
	"github.com/gocolly/colly/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type PlayersAwardsMessage struct {
	PlayerID int64 `json:"player_id"`
}

func (m PlayersAwardsMessage) Queue() rabbit.QueueName {
	return QueuePlayersAwards
}

func playerAwardsHandler(message *rabbit.Message) {

	payload := PlayersAwardsMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToFailQueue(message)
		return
	}

	// Websocket
	defer sendPlayerWebsocket(payload.PlayerID, "award", message)

	c := colly.NewCollector(
		colly.AllowURLRevisit(),
		steam.WithAgeCheckCookie,
		steam.WithTimeout(0),
	)

	var awardsGivenCount int
	var awardsGivenPoints int
	var awardsReceivedCount int
	var awardsReceivedPoints int

	c.OnHTML("div.profile_awards_header_subtitle", func(e *colly.HTMLElement) {

		matches := helpers.RegexIntsCommas.FindAllString(e.Text, -1)
		if len(matches) == 2 {
			if strings.Contains(e.Text, "Given") {

				awardsGivenCount = helpers.StringToInt(matches[0])
				awardsGivenPoints = helpers.StringToInt(matches[1])

			} else if strings.Contains(e.Text, "Received") {

				awardsReceivedCount = helpers.StringToInt(matches[0])
				awardsReceivedPoints = helpers.StringToInt(matches[1])
			}
		}
	})

	var url = "https://steamcommunity.com/profiles/" + strconv.FormatInt(payload.PlayerID, 10) + "/awards/"

	c.OnError(func(r *colly.Response, err error) {
		steam.LogSteamError(err, zap.String("url", url), zap.String("body", string(message.Message.Body)))
	})

	err = c.Visit(url)
	if err != nil {
		// steam.LogSteamError(err) // Already logged
		sendToRetryQueue(message)
		return
	}

	if awardsGivenCount == 0 && awardsGivenPoints == 0 && awardsReceivedCount == 0 && awardsReceivedPoints == 0 {

		message.Ack()
		return
	}

	// Update in Mongo
	var update = bson.D{
		{"awards_given_count", awardsGivenCount},
		{"awards_given_points", awardsGivenPoints},
		{"awards_received_count", awardsReceivedCount},
		{"awards_received_points", awardsReceivedPoints},
	}

	_, err = mongo.UpdateOne(mongo.CollectionPlayers, bson.D{{"_id", payload.PlayerID}}, update)
	if err != nil {
		log.ErrS(err, payload.PlayerID)
		sendToRetryQueue(message)
		return
	}

	err = memcache.Client().Delete(memcache.ItemPlayer(payload.PlayerID).Key)
	if err != nil {
		log.ErrS(err, payload.PlayerID)
		sendToRetryQueue(message)
		return
	}

	// Update in Elastic
	err = ProducePlayerSearch(nil, payload.PlayerID)
	if err != nil {
		log.ErrS(err, payload.PlayerID)
		sendToRetryQueue(message)
		return
	}

	// Add to Influx
	fields := map[schemas.PlayerField]interface{}{
		schemas.InfPlayersAwardsGivenCount:     awardsGivenCount,
		schemas.InfPlayersAwardsGivenPoints:    awardsGivenPoints,
		schemas.InfPlayersAwardsReceivedCount:  awardsReceivedCount,
		schemas.InfPlayersAwardsReceivedPoints: awardsReceivedPoints,
	}

	err = savePlayerStatsToInflux(payload.PlayerID, fields)
	if err != nil {
		log.ErrS(err, payload.PlayerID)
		sendToRetryQueue(message)
		return
	}

	//
	message.Ack()
}
