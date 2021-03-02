package queue

import (
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/steam"
	"github.com/gocolly/colly/v2"
	influx "github.com/influxdata/influxdb1-client"
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

	c.OnError(func(r *colly.Response, err error) {
		steam.LogSteamError(err)
	})

	err = c.Visit("https://steamcommunity.com/profiles/" + strconv.FormatInt(payload.PlayerID, 10) + "/awards/")
	if err != nil {
		steam.LogSteamError(err, zap.String("body", string(message.Message.Body)))
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

	_, err = mongo.UpdateOne(mongo.CollectionPlayers, bson.D{{"_id", payload.PlayerID}}, update, nil)
	if err != nil {
		log.ErrS(err, payload.PlayerID)
		sendToRetryQueue(message)
		return
	}

	err = memcache.Delete(memcache.ItemPlayer(payload.PlayerID).Key)
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
	point := influx.Point{
		Measurement: string(influxHelper.InfluxMeasurementPlayers),
		Tags: map[string]string{
			"player_id": strconv.FormatInt(payload.PlayerID, 10),
		},
		Fields: map[string]interface{}{
			"awards_given_count":     awardsGivenCount,
			"awards_given_points":    awardsGivenPoints,
			"awards_received_count":  awardsReceivedCount,
			"awards_received_points": awardsReceivedPoints,
		},
		Time:      time.Now(),
		Precision: "m",
	}

	_, err = influxHelper.InfluxWrite(influxHelper.InfluxRetentionPolicyAllTime, point)
	if err != nil {
		log.ErrS(err, payload.PlayerID)
		sendToRetryQueue(message)
		return
	}

	//
	message.Ack()
}
