package queue

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/Jleagle/influxql"
	"github.com/Jleagle/steam-go/steam"
	"github.com/cenkalti/backoff"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/sql"
	influx "github.com/influxdata/influxdb1-client"
	"github.com/mitchellh/mapstructure"
	"github.com/nicklaw5/helix"
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
		Message:       appPlayerMessage{},
		OriginalQueue: queueGoAppPlayer,
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

	// Get apps
	appMap := map[int]sql.App{}
	apps, err := sql.GetAppsByID(message.IDs, []string{"id", "twitch_id"})
	if err != nil {
		logError(err)
		payload.ackRetry(msg)
		return
	}

	for _, v := range apps {
		appMap[v.ID] = v
	}

	for _, appID := range message.IDs {

		if payload.Attempt > 1 {
			logInfo("Consuming app player " + strconv.Itoa(appID) + ", attempt " + strconv.Itoa(payload.Attempt))
		}

		app, ok := appMap[appID]
		if ok {

			viewers, err := getAppTwitchStreamers(app.TwitchID)
			if err != nil {
				logError(err, appID)
				payload.ackRetry(msg)
				return
			}

			err = saveAppPlayerToInflux(&app, viewers)
			if err != nil {
				helpers.LogSteamError(err, appID)
				payload.ackRetry(msg)
				return
			}

			err = updateAppPlayerInfoRow(&app)
			if err != nil {
				logError(err, appID)
				payload.ackRetry(msg)
				return
			}
		}
	}

	//
	payload.ack(msg)
}

func getAppTwitchStreamers(twitchID int) (viewers int, err error) {

	if twitchID > 0 {

		client, err := helpers.GetTwitch()
		if err != nil {
			return 0, err
		}

		var resp *helix.StreamsResponse

		// Retrying as this call can fail
		operation := func() (err error) {

			resp, err = client.GetStreams(&helix.StreamsParams{First: 100, GameIDs: []string{strconv.Itoa(twitchID)}, Language: []string{"en"}})
			return err
		}

		policy := backoff.NewExponentialBackOff()

		err = backoff.RetryNotify(operation, backoff.WithMaxRetries(policy, 3), func(err error, t time.Duration) { logInfo(err) })
		if err != nil {
			return 0, err
		}

		for _, v := range resp.Data.Streams {
			viewers += v.ViewerCount
		}
	}

	return viewers, nil
}

var appPlayerSteamClient *steam.Steam

func saveAppPlayerToInflux(app *sql.App, viewers int) (err error) {

	if appPlayerSteamClient == nil {

		appPlayerSteamClient = &steam.Steam{}
		appPlayerSteamClient.SetKey(config.Config.SteamAPIKey.Get())
		appPlayerSteamClient.SetUserAgent("gamedb.online#GetNumberOfCurrentPlayers")
		appPlayerSteamClient.SetAPIRateLimit(time.Millisecond*1100, 10)
	}

	count, b, err := appPlayerSteamClient.GetNumberOfCurrentPlayers(app.ID)
	err = helpers.AllowSteamCodes(err, b, []int{404})
	if err != nil {
		return err
	}

	_, err = helpers.InfluxWrite(helpers.InfluxRetentionPolicyAllTime, influx.Point{
		Measurement: string(helpers.InfluxMeasurementApps),
		Tags: map[string]string{
			"app_id": strconv.Itoa(app.ID),
		},
		Fields: map[string]interface{}{
			"player_count":   count,
			"twitch_viewers": viewers,
		},
		Time:      time.Now(),
		Precision: "m",
	})

	return err
}

func updateAppPlayerInfoRow(app *sql.App) (err error) {

	var builder *influxql.Builder
	var resp *influx.Response

	// Trend value - https://stackoverflow.com/questions/41361734/get-difference-since-30-days-ago-in-influxql-influxdb
	subBuilder := influxql.NewBuilder()
	subBuilder.AddSelect("difference(last(player_count))", "")
	subBuilder.SetFrom(helpers.InfluxGameDB, helpers.InfluxRetentionPolicyAllTime.String(), helpers.InfluxMeasurementApps.String())
	subBuilder.AddWhere("app_id", "=", app.ID)
	subBuilder.AddWhere("time", ">=", "NOW() - 7d")
	subBuilder.AddGroupByTime("1h")

	builder = influxql.NewBuilder()
	builder.AddSelect("cumulative_sum(difference)", "")
	builder.SetFromSubQuery(subBuilder)

	resp, err = helpers.InfluxQuery(builder.String())
	if err != nil {
		return err
	}

	var trendTotal int64

	// Get the last value, todo, put into influx helper, like the ones below
	if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {
		values := resp.Results[0].Series[0].Values
		if len(values) > 0 {

			last := values[len(values)-1]

			trendTotal, err = last[1].(json.Number).Int64()
			if err != nil {
				return err
			}
		}
	}

	// 7 Days
	builder = influxql.NewBuilder()
	builder.AddSelect("max(player_count)", "max_player_count")
	builder.SetFrom(helpers.InfluxGameDB, helpers.InfluxRetentionPolicyAllTime.String(), helpers.InfluxMeasurementApps.String())
	builder.AddWhere("time", ">", "NOW() - 7d")
	builder.AddWhere("app_id", "=", app.ID)
	builder.SetFillNone()

	resp, err = helpers.InfluxQuery(builder.String())
	if err != nil {
		return err
	}

	var week = helpers.GetFirstInfluxInt(resp)

	// All time
	builder = influxql.NewBuilder()
	builder.AddSelect("max(player_count)", "max_player_count")
	builder.SetFrom(helpers.InfluxGameDB, helpers.InfluxRetentionPolicyAllTime.String(), helpers.InfluxMeasurementApps.String())
	builder.AddWhere("app_id", "=", app.ID)
	builder.SetFillNone()

	resp, err = helpers.InfluxQuery(builder.String())
	if err != nil {
		return err
	}

	var alltime = helpers.GetFirstInfluxInt(resp)

	gorm, err := sql.GetMySQLClient()
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"player_trend":        int(trendTotal),
		"player_peak_week":    week,
		"player_peak_alltime": alltime,
	}

	gorm.Table("apps").Where("id = ?", app.ID).Updates(data)

	return gorm.Error
}
