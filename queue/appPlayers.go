package queue

import (
	"strconv"
	"time"

	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	influx "github.com/influxdata/influxdb1-client"
	"github.com/mitchellh/mapstructure"
	"github.com/streadway/amqp"
)

type appPlayerMessage struct {
	ID int `json:"id"`
}

type appPlayerQueue struct {
	baseQueue
}

func (q appPlayerQueue) processMessages(msgs []amqp.Delivery) {

	msg := msgs[0]

	var err error
	var payload = baseMessage{
		Message: appPlayerMessage{},
	}

	err = helpers.Unmarshal(msg.Body, &payload)
	if err != nil {
		logError(err)
		payload.ack(msg)
		return
	}

	var message appPlayerMessage
	err = mapstructure.Decode(payload.Message, &message)
	if err != nil {
		logError(err)
		payload.ack(msg)
		return
	}

	if payload.Attempt > 1 {
		logInfo("Consuming app player " + strconv.Itoa(message.ID) + ", attempt " + strconv.Itoa(payload.Attempt))
	}

	err = saveAppPlayerToInflux(message.ID)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	//
	payload.ack(msg)
}

func saveAppPlayerToInflux(appID int) (err error) {

	s := helpers.GetSteam()
	sx := *s
	sx.SetAPIRateLimit(time.Millisecond*600, 10)
	count, _, err := sx.GetNumberOfCurrentPlayers(appID)
	if err != nil {
		return err
	}

	_, err = db.InfluxWrite(db.InfluxRetentionPolicyAllTime, influx.Point{
		Measurement: string(db.InfluxMeasurementApps),
		Tags: map[string]string{
			"app_id": strconv.Itoa(appID),
		},
		Fields: map[string]interface{}{
			"player_count": count,
		},
		Time:      time.Now(),
		Precision: "m",
	})

	return err
}
