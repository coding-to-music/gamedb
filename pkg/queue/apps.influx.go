package queue

import (
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"github.com/Jleagle/influxql"
	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	mongoHelper "github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AppInfluxMessage struct {
	AppIDs []int `json:"app_ids"`
}

func (m AppInfluxMessage) Queue() rabbit.QueueName {
	return QueueAppsInflux
}

func appInfluxHandler(message *rabbit.Message) {

	// Sleep to not cause influx memory to spike too much
	time.Sleep(time.Second * 5)

	payload := AppInfluxMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err, message.Message.Body)
		sendToFailQueue(message)
		return
	}

	if len(payload.AppIDs) == 0 {
		message.Ack()
		return
	}

	var wg sync.WaitGroup

	wg.Add(1)
	var appPlayersWeek = map[int]int64{}
	go func() {

		defer wg.Done()

		var err error
		appPlayersWeek, err = getAppTopPlayersWeek(payload.AppIDs)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToRetryQueue(message)
			return
		}
	}()

	// wg.Add(1)
	// var appPlayersWeekAverage map[int]float64
	// go func() {
	//
	// 	defer wg.Done()
	//
	// 	var err error
	// 	appPlayersWeekAverage, err = getAppAveragePlayersWeek(payload.AppIDs)
	// 	if err != nil {
	// 		log.Err(err, message.Message.Body)
	// 		sendToRetryQueue(message)
	// 		return
	// 	}
	// }()

	wg.Add(1)
	var appPlayersAlltime map[int]int64
	go func() {

		defer wg.Done()

		var err error
		appPlayersAlltime, err = getAppTopPlayersAlltime(payload.AppIDs)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToRetryQueue(message)
			return
		}
	}()

	wg.Add(1)
	var appTrend map[int]float64
	go func() {

		defer wg.Done()

		var err error
		appTrend, err = getAppTrendValue(payload.AppIDs)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToRetryQueue(message)
			return
		}
	}()

	wg.Wait()

	if message.ActionTaken {
		return
	}

	if len(payload.AppIDs) == 0 {
		message.Ack()
		return
	}

	// Save to Mongo
	var writes []mongo.WriteModel
	for _, appID := range payload.AppIDs {

		update := bson.M{}

		if val, ok := appTrend[appID]; ok {
			update["player_trend"] = val
		}

		if val, ok := appPlayersWeek[appID]; ok {
			update["player_peak_week"] = val
		}

		if val, ok := appPlayersAlltime[appID]; ok {
			update["player_peak_alltime"] = val
		}

		// if val, ok := appPlayersWeekAverage[appID]; ok {
		// 	update["player_avg_week"] = val
		// }

		if len(update) > 0 {

			write := mongo.NewUpdateOneModel()
			write.SetFilter(bson.M{"_id": appID})
			write.SetUpdate(bson.M{"$set": update})
			write.SetUpsert(false)

			writes = append(writes, write)
		}
	}

	err = mongoHelper.UpdateAppsInflux(writes)
	if err != nil {
		log.Err(err, message.Message.Body)
		sendToRetryQueue(message)
		return
	}

	// Clear app cache
	var items []string
	for _, v := range payload.AppIDs {
		items = append(items, memcache.MemcacheApp(v).Key)
	}

	err = memcache.Delete(items...)
	if err != nil {
		log.Err(err, message.Message.Body)
		sendToRetryQueue(message)
		return
	}

	// Update in Elastic
	for _, v := range payload.AppIDs {
		err = ProduceAppSearch(nil, v)
		log.Err(err)
	}

	//
	message.Ack()
}

func getAppTopPlayersWeek(appIDs []int) (vals map[int]int64, err error) {

	builder := influxql.NewBuilder()
	builder.AddSelect("max(player_count)", "max_player_count")
	builder.SetFrom(influxHelper.InfluxGameDB, influxHelper.InfluxRetentionPolicyAllTime.String(), influxHelper.InfluxMeasurementApps.String())
	builder.AddWhere("time", ">", "NOW() - 7d")
	builder.AddWhereRaw(`"app_id" =~ /^(` + helpers.JoinInts(appIDs, "|") + `)$/`)
	builder.AddGroupBy("app_id")
	builder.SetFillNumber(0)

	resp, err := influxHelper.InfluxQuery(builder.String())
	if err != nil {
		return vals, err
	}

	vals = map[int]int64{}
	for _, v := range resp.Results[0].Series {

		appID, err := strconv.Atoi(v.Tags["app_id"])
		if err != nil {
			log.Err(err)
			continue
		}

		val, err := v.Values[0][1].(json.Number).Int64()
		if err != nil {
			log.Err(err)
			continue
		}

		vals[appID] = val
	}

	return vals, err
}

func getAppAveragePlayersWeek(appIDs []int) (vals map[int]float64, err error) {

	builder := influxql.NewBuilder()
	builder.AddSelect("mean(player_count)", "mean_player_count")
	builder.SetFrom(influxHelper.InfluxGameDB, influxHelper.InfluxRetentionPolicyAllTime.String(), influxHelper.InfluxMeasurementApps.String())
	builder.AddWhere("time", ">", "NOW() - 7d")
	builder.AddWhereRaw(`"app_id" =~ /^(` + helpers.JoinInts(appIDs, "|") + `)$/`)
	builder.AddGroupBy("app_id")
	builder.SetFillNumber(0)

	resp, err := influxHelper.InfluxQuery(builder.String())
	if err != nil {
		return vals, err
	}

	vals = map[int]float64{}
	for _, v := range resp.Results[0].Series {

		appID, err := strconv.Atoi(v.Tags["app_id"])
		if err != nil {
			log.Err(err)
			continue
		}

		val, err := v.Values[0][1].(json.Number).Float64()
		if err != nil {
			log.Err(err)
			continue
		}

		vals[appID] = val
	}

	return vals, err
}

func getAppTopPlayersAlltime(appIDs []int) (vals map[int]int64, err error) {

	builder := influxql.NewBuilder()
	builder.AddSelect("max(player_count)", "max_player_count")
	builder.SetFrom(influxHelper.InfluxGameDB, influxHelper.InfluxRetentionPolicyAllTime.String(), influxHelper.InfluxMeasurementApps.String())
	builder.AddWhereRaw(`"app_id" =~ /^(` + helpers.JoinInts(appIDs, "|") + `)$/`)
	builder.AddGroupBy("app_id")
	builder.SetFillNumber(0)

	resp, err := influxHelper.InfluxQuery(builder.String())
	if err != nil {
		return vals, err
	}

	vals = map[int]int64{}
	for _, v := range resp.Results[0].Series {

		appID, err := strconv.Atoi(v.Tags["app_id"])
		if err != nil {
			log.Err(err)
			continue
		}

		val, err := v.Values[0][1].(json.Number).Int64()
		if err != nil {
			log.Err(err)
			continue
		}

		vals[appID] = val
	}

	return vals, err
}

func getAppTrendValue(appIDs []int) (vals map[int]float64, err error) {

	builder := influxql.NewBuilder()
	builder.AddSelect("max(player_count)", "max_player_count")
	builder.SetFrom(influxHelper.InfluxGameDB, influxHelper.InfluxRetentionPolicyAllTime.String(), influxHelper.InfluxMeasurementApps.String())
	builder.AddWhere("time", ">", "NOW() - 7d - 1h")
	builder.AddWhereRaw(`"app_id" =~ /^(` + helpers.JoinInts(appIDs, "|") + `)$/`)
	builder.AddGroupByTime("1h")
	builder.AddGroupBy("app_id")
	builder.SetFillNumber(0)

	resp, err := influxHelper.InfluxQuery(builder.String())
	if err != nil {
		return vals, err
	}

	vals = map[int]float64{}

	if len(resp.Results) > 0 {
		for _, v := range resp.Results[0].Series {

			appID, err := strconv.Atoi(v.Tags["app_id"])
			if err != nil {
				log.Err(err)
				continue
			}

			vals[appID] = influxHelper.GetInfluxTrendFromSeries(v, 0)
		}
	}

	return vals, err
}
