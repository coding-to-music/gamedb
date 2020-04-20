package queue

import (
	"encoding/json"
	"sync"

	"github.com/Jleagle/influxql"
	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/helpers/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type AppInfluxMessage struct {
	ID int `json:"id"`
}

func appInfluxHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		payload := AppInfluxMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			continue
		}

		var wg sync.WaitGroup

		wg.Add(1)
		var appPlayersWeek int64
		go func() {

			defer wg.Done()

			var err error
			appPlayersWeek, err = getAppTopPlayersWeek(payload.ID)
			if err != nil {
				log.Err(err, payload.ID)
				sendToRetryQueue(message)
				return
			}
		}()

		wg.Add(1)
		var appPlayersWeekAverage float64
		go func() {

			defer wg.Done()

			var err error
			appPlayersWeekAverage, err = getAppAveragePlayersWeek(payload.ID)
			if err != nil {
				log.Err(err, payload.ID)
				sendToRetryQueue(message)
				return
			}
		}()

		wg.Add(1)
		var appPlayersAlltime int64
		go func() {

			defer wg.Done()

			var err error
			appPlayersAlltime, err = getAppTopPlayersAlltime(payload.ID)
			if err != nil {
				log.Err(err, payload.ID)
				sendToRetryQueue(message)
				return
			}
		}()

		wg.Add(1)
		var appTrend int64
		go func() {

			defer wg.Done()

			var err error
			appTrend, err = getAppTrendValue(payload.ID)
			if err != nil {
				log.Err(err, payload.ID)
				sendToRetryQueue(message)
				return
			}
		}()

		wg.Wait()

		if message.ActionTaken {
			continue
		}

		// Save to Mongo
		wg.Add(1)
		go func() {

			defer wg.Done()

			err = appsInfluxSave(payload.ID, appTrend, appPlayersWeek, appPlayersAlltime, appPlayersWeekAverage)
			if err != nil {
				log.Err(err, payload.ID)
				sendToRetryQueue(message)
				return
			}
		}()

		wg.Wait()

		if message.ActionTaken {
			continue
		}

		message.Ack(false)
	}
}

func getAppTopPlayersWeek(appID int) (val int64, err error) {

	builder := influxql.NewBuilder()
	builder.AddSelect("max(player_count)", "max_player_count")
	builder.SetFrom(influxHelper.InfluxGameDB, influxHelper.InfluxRetentionPolicyAllTime.String(), influxHelper.InfluxMeasurementApps.String())
	builder.AddWhere("time", ">", "NOW() - 7d")
	builder.AddWhere("app_id", "=", appID)
	builder.SetFillNone()

	resp, err := influxHelper.InfluxQuery(builder.String())
	if err != nil {
		return 0, err
	}

	return influxHelper.GetFirstInfluxInt(resp), nil
}

func getAppAveragePlayersWeek(appID int) (val float64, err error) {

	builder := influxql.NewBuilder()
	builder.AddSelect("mean(player_count)", "mean_player_count")
	builder.SetFrom(influxHelper.InfluxGameDB, influxHelper.InfluxRetentionPolicyAllTime.String(), influxHelper.InfluxMeasurementApps.String())
	builder.AddWhere("time", ">", "NOW() - 7d")
	builder.AddWhere("app_id", "=", appID)

	resp, err := influxHelper.InfluxQuery(builder.String())
	if err != nil {
		return 0, err
	}

	return influxHelper.GetFirstInfluxFloat(resp), nil
}

func getAppTopPlayersAlltime(appID int) (val int64, err error) {

	builder := influxql.NewBuilder()
	builder.AddSelect("max(player_count)", "max_player_count")
	builder.SetFrom(influxHelper.InfluxGameDB, influxHelper.InfluxRetentionPolicyAllTime.String(), influxHelper.InfluxMeasurementApps.String())
	builder.AddWhere("app_id", "=", appID)
	builder.SetFillNone()

	resp, err := influxHelper.InfluxQuery(builder.String())
	if err != nil {
		return 0, err
	}

	return influxHelper.GetFirstInfluxInt(resp), nil
}

func getAppTrendValue(appID int) (trend int64, err error) {

	// Trend value - https://stackoverflow.com/questions/41361734/get-difference-since-30-days-ago-in-influxql-influxdb
	subBuilder := influxql.NewBuilder()
	subBuilder.AddSelect("difference(last(player_count))", "")
	subBuilder.SetFrom(influxHelper.InfluxGameDB, influxHelper.InfluxRetentionPolicyAllTime.String(), influxHelper.InfluxMeasurementApps.String())
	subBuilder.AddWhere("app_id", "=", appID)
	subBuilder.AddWhere("time", ">=", "NOW() - 7d")
	subBuilder.AddGroupByTime("1h")

	builder := influxql.NewBuilder()
	builder.AddSelect("cumulative_sum(difference)", "")
	builder.SetFromSubQuery(subBuilder)

	resp, err := influxHelper.InfluxQuery(builder.String())
	if err != nil {
		return 0, err
	}

	var trendTotal int64

	// Get the last value, todo, put into influx helper, like the ones below
	if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {
		values := resp.Results[0].Series[0].Values
		if len(values) > 0 {

			last := values[len(values)-1]

			trendTotal, err = last[1].(json.Number).Int64()
			if err != nil {
				return 0, err
			}
		}
	}

	return trendTotal, nil
}

func appsInfluxSave(appID int, trend int64, week int64, alltime int64, average float64) (err error) {

	_, err = mongo.UpdateOne(mongo.CollectionApps, bson.D{{"_id", appID}}, bson.D{
		{"player_trend", trend},
		{"player_peak_week", week},
		{"player_peak_alltime", alltime},
		{"player_avg_week", average},
	})
	return err
}
