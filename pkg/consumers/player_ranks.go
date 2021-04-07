package consumers

import (
	"strconv"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	influx "github.com/influxdata/influxdb1-client"
	"go.mongodb.org/mongo-driver/bson"
	mongodb "go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

const batchSize = 10000

type PlayerRanksMessage struct {
	ObjectKey  string  `json:"object_key"`
	SortColumn string  `json:"sort_column"`
	Continent  *string `json:"continent"`
	Country    *string `json:"country"`
	State      *string `json:"state"`
}

func playerRanksHandler(message *rabbit.Message) {

	payload := PlayerRanksMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToFailQueue(message)
		return
	}

	if payload.ObjectKey == "" || payload.SortColumn == "" {
		sendToFailQueue(message)
		return
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
	filter = append(filter, bson.E{Key: payload.SortColumn, Value: bson.M{"$gt": 0}})

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
			if len(payload.ObjectKey) == 1 { // Global

				var points []influx.Point
				for position, player := range players {
					if val, ok := helpers.PlayerRankFieldsInflux[helpers.RankMetric(payload.ObjectKey)]; ok {
						points = append(points, influx.Point{
							Measurement: string(influxHelper.InfluxMeasurementPlayers),
							Tags: map[string]string{
								"player_id": strconv.FormatInt(player.ID, 10),
							},
							Fields: map[string]interface{}{
								string(val): offset + int64(position) + 1,
							},
							Time:      time.Now(),
							Precision: "h",
						})
					}
				}

				batch := influx.BatchPoints{
					Points:          points,
					Database:        influxHelper.InfluxGameDB,
					RetentionPolicy: influxHelper.InfluxRetentionPolicyAllTime.String(),
				}

				_, err = influxHelper.InfluxWriteMany(influxHelper.InfluxRetentionPolicyAllTime, batch)
				if err != nil {
					return err
				}
			}

			return nil
		}()

		// Error, add to retry queue and bail
		if err != nil {
			if val, ok := err.(mongodb.BulkWriteException); ok {
				for _, err2 := range val.WriteErrors {
					log.ErrS(err2, err2.Request)
				}
			} else {
				log.ErrS(err, payload)
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

	message.Ack()
}
