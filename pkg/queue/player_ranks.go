package queue

import (
	"strconv"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	influx "github.com/influxdata/influxdb1-client"
	"go.mongodb.org/mongo-driver/bson"
	mongodb "go.mongodb.org/mongo-driver/mongo"
)

const (
	batchSize   = 20000
	addToInflux = 1000
)

type PlayerRanksMessage struct {
	ObjectKey  string  `json:"object_key"`
	SortColumn string  `json:"sort_column"`
	Continent  *string `json:"continent"`
	Country    *string `json:"country"`
	State      *string `json:"state"`
}

func playerRanksHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		payload := PlayerRanksMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			continue
		}

		if payload.ObjectKey == "" || payload.SortColumn == "" {
			sendToFailQueue(message)
			continue
		}

		// todo, remove when we have more memory
		if config.IsProd() && payload.Country == nil && payload.State == nil {
			sendToRetryQueue(message)
			continue
		}

		// Create filter
		var filter = bson.D{}
		if payload.Continent != nil {
			filter = append(filter, bson.E{Key: "continent_code", Value: *payload.Continent})
		}
		if payload.Country != nil {
			filter = append(filter, bson.E{Key: "country_code", Value: *payload.Country})
		}
		if payload.State != nil {
			filter = append(filter, bson.E{Key: "status_code", Value: *payload.State})
		}
		filter = append(filter, bson.E{Key: payload.SortColumn, Value: bson.M{"$exists": true, "$gt": 0}}) // Put last to help indexes

		// Batched to use less memory consumer memory
		var offset int64
		var players []mongo.Player

		for {

			err := func() error {

				// Get players
				players, err = mongo.GetPlayers(offset, batchSize, bson.D{{payload.SortColumn, -1}}, filter, bson.M{"_id": 1})
				if err != nil {
					return err
				}

				// Update players
				var writes []mongodb.WriteModel
				for position, player := range players {

					write := mongodb.NewUpdateOneModel()
					write.SetFilter(bson.M{"_id": player.ID})
					write.SetUpdate(bson.M{"$set": bson.M{"ranks." + payload.ObjectKey: offset + int64(position) + 1}})
					write.SetUpsert(true)

					writes = append(writes, write)
				}

				err = mongo.BulkUpdatePlayers(writes)
				if err != nil {
					return err
				}

				// Add player ranks to Influx
				var points []influx.Point
				if len(payload.ObjectKey) == 1 { // Global
					for position, player := range players {
						if val, ok := mongo.PlayerRankFieldsInflux[mongo.RankMetric(payload.ObjectKey)]; ok {
							points = append(points, influx.Point{
								Measurement: string(influxHelper.InfluxMeasurementAPICalls),
								Tags: map[string]string{
									"player_id": strconv.FormatInt(player.ID, 10),
								},
								Fields: map[string]interface{}{
									val: offset + int64(position) + 1,
								},
								Time:      time.Now(),
								Precision: "h",
							})
						}
					}
				}

				_, err = influxHelper.InfluxWriteMany(influxHelper.InfluxRetentionPolicyAllTime, influx.BatchPoints{
					Points:          points,
					Database:        influxHelper.InfluxGameDB,
					RetentionPolicy: influxHelper.InfluxRetentionPolicyAllTime.String(),
					Precision:       "m",
				})

				return err
			}()

			// Error, add to retry queue and bail
			if err != nil {
				if val, ok := err.(mongodb.BulkWriteException); ok {
					for _, err2 := range val.WriteErrors {
						log.Err(err2, err2.Request)
					}
				} else {
					log.Err(err)
				}

				sendToRetryQueue(message)
				break
			}

			// Last batch, bail
			if len(players) < batchSize {
				break
			}

			offset += batchSize
		}

		message.Ack(false)
	}
}
