package queue

import (
	"strconv"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	influx "github.com/influxdata/influxdb1-client"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type StatsMessage struct {
	Type      mongo.StatsType `json:"type"`
	StatID    int             `json:"id"`
	AppsCount int64           `json:"apps_count"`
}

func (m StatsMessage) Queue() rabbit.QueueName {
	return QueueStats
}

func statsHandler(message *rabbit.Message) {

	if !config.IsLocal() {
		message.Ack()
		return
	}

	payload := StatsMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err.Error(), zap.ByteString("message", message.Message.Body))
		sendToFailQueue(message)
		return
	}

	if payload.AppsCount == 0 {
		log.Err("Missing app count", zap.ByteString("message", message.Message.Body))
		sendToRetryQueue(message)
		return
	}

	var totalApps int
	var totalScore float64
	var totalPrice = map[steamapi.ProductCC]int{}
	var totalPlayers int

	err = mongo.BatchApps(bson.D{{payload.Type.MongoCol(), payload.StatID}}, nil, func(apps []mongo.App) {

		for _, app := range apps {

			// Count
			totalApps++

			// Score
			totalScore += app.ReviewsScore

			// Prices
			for k, v := range app.Prices {
				totalPrice[k] += v.Final
			}

			// Players
			totalPlayers += app.PlayerPeakWeek
		}
	})

	var meanScore float64
	var meanPlayers float64
	var meanPrice = map[steamapi.ProductCC]float32{}

	if totalApps > 0 {

		meanScore = totalScore / float64(totalApps)
		meanPlayers = float64(totalPlayers) / float64(totalApps)

		for k, v := range totalPrice {
			meanPrice[k] = float32(v) / float32(totalApps)
		}
	}

	// Update Mongo
	filter := bson.D{
		{"type", payload.Type},
		{"id", payload.StatID},
	}
	update := bson.D{
		{Key: "apps", Value: totalApps},
		{Key: "mean_price", Value: meanPrice},
		{Key: "mean_score", Value: meanScore},
		{Key: "mean_players", Value: meanPlayers},
	}

	_, err = mongo.UpdateOne(mongo.CollectionStats, filter, update)
	if err != nil {
		log.Err(err.Error(), zap.ByteString("message", message.Message.Body))
		sendToRetryQueue(message)
		return
	}

	// Update Influx
	fields := map[string]interface{}{
		"app_count":    totalApps,
		"percent":      (float64(totalApps) / float64(payload.AppsCount)) * 100,
		"mean_price":   meanPrice,
		"mean_score":   meanScore,
		"mean_players": meanPlayers,
	}

	point := influx.Point{
		Measurement: string(influxHelper.InfluxMeasurementStats),
		Tags: map[string]string{
			"type": string(payload.Type),
			"id":   strconv.Itoa(payload.StatID),
		},
		Fields:    fields,
		Time:      time.Now(),
		Precision: "h",
	}

	_, err = influxHelper.InfluxWrite(influxHelper.InfluxRetentionPolicyAllTime, point)
	if err != nil {
		log.Err(err.Error(), zap.ByteString("message", message.Message.Body))
		sendToRetryQueue(message)
		return
	}

	//
	message.Ack()
}
