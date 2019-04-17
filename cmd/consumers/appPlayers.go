package main

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/Jleagle/influxql"
	"github.com/Jleagle/steam-go/steam"
	"github.com/cenkalti/backoff"
	"github.com/gamedb/website/pkg"
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

	// Get apps
	appMap := map[int]sql.App{}
	apps, err := pkg.GetAppsByID(message.IDs, []string{"id", "twitch_id"})
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
		if ok || appID == 0 {

			if appID == 0 {
				app = sql.App{}
			}

			viewers, err := getAppTwitchStreamers(&app)
			if err != nil {
				logError(err, appID)
				payload.ackRetry(msg)
				return
			}

			err = saveAppPlayerToInflux(&app, viewers)
			if err != nil {
				logError(err, appID)
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

func getAppTwitchStreamers(app *sql.App) (viewers int, err error) {

	if app.TwitchID > 0 {

		client, err := pkg.GetTwitch()
		if err != nil {
			return 0, err
		}

		var resp *helix.StreamsResponse

		// Retrying as this call can fail
		operation := func() (err error) {

			resp, err = client.GetStreams(&helix.StreamsParams{First: 100, GameIDs: []string{strconv.Itoa(app.TwitchID)}, Language: []string{"en"}})
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
		// appPlayerSteamClient.SetAPIRateLimit(time.Millisecond*1000, 10)
		// appPlayerSteamClient.SetStoreRateLimit(time.Millisecond*1600, 10)
	}

	count, b, err := appPlayerSteamClient.GetNumberOfCurrentPlayers(app.ID)
	err = helpers.HandleSteamStoreErr(err, b, []int{404})
	if err != nil {
		return err
	}

	_, err = pkg.InfluxWrite(pkg.InfluxRetentionPolicyAllTime, influx.Point{
		Measurement: string(pkg.InfluxMeasurementApps),
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

	if app == nil || app.ID == 0 {
		return nil
	}

	var resp *influx.Response

	// Trend value - https://stackoverflow.com/questions/41361734/get-difference-since-30-days-ago-in-influxql-influxdb
	query := `SELECT cumulative_sum(difference) FROM (
		SELECT difference(last("player_count")) FROM "GameDB"."alltime"."apps" WHERE "app_id" = '` + strconv.Itoa(app.ID) + `' AND time >= now() - 7d GROUP BY time(1h)
	)`

	resp, err = pkg.InfluxQuery(query)
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

	var builder *influxql.Builder

	// 7 Days
	builder = influxql.NewBuilder()
	builder.AddSelect("max(player_count)", "max_player_count")
	builder.SetFrom("GameDB", "alltime", "apps")
	builder.AddWhere("time", ">", "NOW() - 7d")
	builder.AddWhere("app_id", "=", app.ID)
	builder.SetFillNone()

	resp, err = pkg.InfluxQuery(builder.String())
	if err != nil {
		return err
	}

	var week = pkg.GetFirstInfluxInt(resp)

	// All time
	builder = influxql.NewBuilder()
	builder.AddSelect("max(player_count)", "max_player_count")
	builder.SetFrom("GameDB", "alltime", "apps")
	builder.AddWhere("app_id", "=", app.ID)
	builder.SetFillNone()

	resp, err = pkg.InfluxQuery(builder.String())
	if err != nil {
		return err
	}

	var alltime = pkg.GetFirstInfluxInt(resp)

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
