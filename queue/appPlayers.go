package queue

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	influx "github.com/influxdata/influxdb1-client"
	"github.com/mitchellh/mapstructure"
	"github.com/streadway/amqp"
)

type appPlayerMessage struct {
	IDs []int `json:"ids"`
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

	for _, appID := range message.IDs {

		if payload.Attempt > 1 {
			logInfo("Consuming app player " + strconv.Itoa(appID) + ", attempt " + strconv.Itoa(payload.Attempt))
		}

		err = saveAppPlayerToInflux(appID)
		if err != nil {
			logError(err, appID)
			payload.ackRetry(msg)
			return
		}
	}

	//
	payload.ack(msg)
}

func saveAppPlayerToInflux(appID int) (err error) {

	s := helpers.GetSteam()
	sx := *s
	sx.SetAPIRateLimit(time.Millisecond*600, 10)
	count, _, err := sx.GetNumberOfCurrentPlayers(appID)

	steamErr, ok := err.(steam.Error)
	if ok && (steamErr.Code == 404) {
		err = nil
	}
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

func UpdateAppRow(appID int) {

	query := `SELECT cumulative_sum(difference) FROM (
		SELECT difference(last("player_count")) FROM "GameDB"."alltime"."apps" WHERE "app_id" = '` + strconv.Itoa(appID) + `' AND time >= now() - 7d GROUP BY time(1h)
	)`

	resp, err := db.InfluxQuery(query)
	if err != nil {
		log.Err(err)
		return
	}

	values := resp.Results[0].Series[0].Values
	if len(values) > 0 {

		last := values[len(values)-1]

		fmt.Println(last[1])
	}

	gorm, err := db.GetMySQLClient()
	if err != nil {
		log.Err(err)
		return
	}

	app := db.App{}
	app.ID = appID

	gorm = gorm.Model(&app).UpdateColumns(db.App{PlayerTrend: 1})
	log.Err(gorm.Error)
}
