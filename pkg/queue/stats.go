package queue

import (
	"math"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	influx "github.com/influxdata/influxdb1-client"
	"github.com/montanaflynn/stats"
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

	payload := StatsMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToFailQueue(message)
		return
	}

	stat := mongo.Stat{}
	stat.Type = payload.Type
	stat.ID = payload.StatID

	if payload.AppsCount == 0 {
		log.Err("Missing app count", zap.String("body", string(message.Message.Body)))
		sendToRetryQueue(message)
		return
	}

	var totalApps int
	var scores stats.Float64Data
	var prices = map[steamapi.ProductCC]stats.Float64Data{}
	var players stats.Float64Data

	filter := bson.D{{payload.Type.MongoCol(), payload.StatID}}
	projection := bson.M{"reviews_score": 1, "prices": 1, "player_peak_week": 1}

	callback := func(apps []mongo.App) {

		for _, app := range apps {

			// Counts
			totalApps++

			// Score
			if app.ReviewsScore > 0 {
				scores = append(scores, app.ReviewsScore)
			}

			// Prices
			for k, v := range app.Prices {
				prices[k] = append(prices[k], float64(v.Final))
			}

			// Players
			players = append(players, float64(app.PlayerPeakWeek))
		}
	}

	err = mongo.BatchApps(filter, projection, callback)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToRetryQueue(message)
		return
	}

	var percent = (float64(totalApps) / float64(payload.AppsCount)) * 100

	// Calculate means
	meanScore, _ := scores.Mean()
	if math.IsNaN(meanScore) {
		meanScore = 0
	}

	meanPlayers, _ := players.Mean()
	if math.IsNaN(meanPlayers) {
		meanPlayers = 0
	}

	meanPrice := map[steamapi.ProductCC]float32{}
	for k, v := range prices {
		f, _ := v.Mean()
		meanPrice[k] = float32(f)
	}

	// Calculate medians
	medianScore, _ := scores.Median()
	if math.IsNaN(medianScore) {
		medianScore = 0
	}

	medianPlayers, _ := players.Median()
	if math.IsNaN(medianPlayers) {
		medianPlayers = 0
	}

	medianPrice := map[steamapi.ProductCC]int{}
	for k, v := range prices {
		f, _ := v.Median()
		medianPrice[k] = int(f)
	}

	// Update Mongo
	update := bson.D{
		{Key: "apps", Value: totalApps},
		{Key: "apps_percent", Value: float32(percent)},

		{Key: "mean_price", Value: meanPrice},
		{Key: "mean_score", Value: float32(meanScore)},
		{Key: "mean_players", Value: meanPlayers},

		{Key: "median_price", Value: medianPrice},
		{Key: "median_score", Value: float32(medianScore)},
		{Key: "median_players", Value: int(medianPlayers)},
	}

	_, err = mongo.UpdateOne(mongo.CollectionStats, bson.D{{"_id", stat.GetKey()}}, update)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToRetryQueue(message)
		return
	}

	// Update Influx
	fields := map[string]interface{}{
		"apps_count":     totalApps,
		"apps_percent":   percent,
		"mean_score":     float32(meanScore),
		"mean_players":   meanPlayers,
		"median_score":   float32(medianScore),
		"median_players": int(medianPlayers),
	}

	for k, v := range meanPrice {
		fields["mean_price_"+string(k)] = v
	}

	for k, v := range medianPrice {
		fields["median_price_"+string(k)] = v
	}

	point := influx.Point{
		Measurement: string(influxHelper.InfluxMeasurementStats),
		Tags: map[string]string{
			"key": stat.GetKey(),
		},
		Fields:    fields,
		Time:      time.Now(),
		Precision: "h",
	}

	_, err = influxHelper.InfluxWrite(influxHelper.InfluxRetentionPolicyAllTime, point)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToRetryQueue(message)
		return
	}

	//
	message.Ack()
}
