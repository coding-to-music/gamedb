package queue

import (
	"encoding/json"
	"math"
	"sync"
	"time"

	"github.com/Jleagle/influxql"
	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/helpers/influx"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"gonum.org/v1/gonum/stat"
)

type AppInfluxMessage struct {
	ID int `json:"id"`
}

func appInfluxHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		// Sleep to not cause influx memory to spike too much
		time.Sleep(time.Second / 5)

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
		_, err = mongo.UpdateOne(mongo.CollectionApps, bson.D{{"_id", payload.ID}}, bson.D{
			{"player_trend", appTrend},
			{"player_peak_week", appPlayersWeek},
			{"player_peak_alltime", appPlayersAlltime},
			{"player_avg_week", appPlayersWeekAverage},
		})
		if err != nil {
			log.Err(err, payload.ID)
			sendToRetryQueue(message)
			continue
		}

		// Clear app cache
		err = memcache.Delete(memcache.MemcacheApp(payload.ID).Key)
		if err != nil {
			log.Err(err, payload.ID)
			sendToRetryQueue(message)
			continue
		}

		//
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
	builder.SetFillNone()

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

	builder := influxql.NewBuilder()
	builder.AddSelect("max(player_count)", "max_player_count")
	builder.SetFrom(influxHelper.InfluxGameDB, influxHelper.InfluxRetentionPolicyAllTime.String(), influxHelper.InfluxMeasurementApps.String())
	builder.AddWhere("time", ">", "NOW() - 7d")
	builder.AddWhere("app_id", "=", appID)
	builder.AddGroupByTime("1h")
	builder.SetFillNone()

	resp, err := influxHelper.InfluxQuery(builder.String())
	if err != nil {
		return 0, err
	}

	var xs []float64
	var ys []float64

	if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 && len(resp.Results[0].Series[0].Values) > 0 {
		for _, v := range resp.Results[0].Series[0].Values {

			trendTotal, err := v[1].(json.Number).Int64()
			if err != nil {
				log.Err(err)
				continue
			}

			t, err := time.Parse(time.RFC3339, v[0].(string))
			if err != nil {
				log.Err(err)
				continue
			}

			xs = append(xs, float64(t.Unix()/60/60)) // Divide to get hours, not seconds
			ys = append(ys, float64(trendTotal))
		}
	} else {
		return 0, nil
	}

	_, slope := stat.LinearRegression(xs, ys, nil, false)
	deg := math.Atan(slope) * (180.0 / math.Pi)
	trend = int64(math.Round(deg * 1000))

	return trend, nil
}
