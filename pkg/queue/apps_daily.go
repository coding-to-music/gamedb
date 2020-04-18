package queue

import (
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Jleagle/influxql"
	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/helpers/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	influx "github.com/influxdata/influxdb1-client"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/api/googleapi"
)

type AppDailyMessage struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	TopApp bool   `json:"top_app"`
}

func appDailyHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		payload := AppDailyMessage{}

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

		wg.Add(1)
		var youtubeViews uint64
		var youtubeComments uint64
		go func() {

			defer wg.Done()

			var err error
			youtubeViews, youtubeComments, err = getYouTubeStats(payload)
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

			err = appsDailySaveToMongo(payload.ID, appTrend, appPlayersWeek, appPlayersAlltime, appPlayersWeekAverage)
			if err != nil {
				log.Err(err, payload.ID)
				sendToRetryQueue(message)
				return
			}
		}()

		// Save to Influx
		wg.Add(1)
		go func() {

			defer wg.Done()

			err = appsDailySaveToInflux(payload, youtubeViews, youtubeComments)
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

func getYouTubeStats(payload AppDailyMessage) (uint64, uint64, error) {

	if !payload.TopApp {
		return 0, 0, nil
	}

	payload.Name = strings.TrimSpace(payload.Name)
	if payload.Name == "" {
		return 0, 0, nil
	}

	// `part` can be:
	// id, snippet, contentDetails, fileDetails, player, processingDetails, recordingDetails, statistics, status, suggestions, topicDetails

	// Get video IDs from search
	searchResponse, err := helpers.YoutubeService.Search.List("id").
		Context(helpers.YoutubeContext).
		MaxResults(50).
		SafeSearch("none").
		Type("video").
		Q(payload.Name).
		Order("viewCount").
		PublishedAfter(time.Now().Add(-time.Hour * 24).Format(time.RFC3339)).
		Do()

	if err != nil {
		return 0, 0, err
	}

	var ids []string
	for _, v := range searchResponse.Items {
		ids = append(ids, v.Id.VideoId)
	}

	// Get video statistics from IDs
	listResponse, err := helpers.YoutubeService.Videos.
		List("statistics").
		Id(strings.Join(ids, ",")).
		Do()

	if err != nil {
		return 0, 0, err
	}

	var views uint64
	var comments uint64
	for _, v := range listResponse.Items {
		views += v.Statistics.ViewCount
		comments += v.Statistics.CommentCount
	}

	//
	return views, comments, nil
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

func appsDailySaveToMongo(appID int, trend int64, week int64, alltime int64, average float64) (err error) {

	_, err = mongo.UpdateOne(mongo.CollectionApps, bson.D{{"_id", appID}}, bson.D{
		{"player_trend", trend},
		{"player_peak_week", week},
		{"player_peak_alltime", alltime},
		{"player_avg_week", average},
	})
	return err
}

func appsDailySaveToInflux(payload AppDailyMessage, views uint64, comments uint64) (err error) {

	if views > 0 || comments > 0 {

		_, err = influxHelper.InfluxWrite(influxHelper.InfluxRetentionPolicyAllTime, influx.Point{
			Measurement: string(influxHelper.InfluxMeasurementApps),
			Tags: map[string]string{
				"app_id": strconv.Itoa(payload.ID),
			},
			Fields: map[string]interface{}{
				"youtube_views":    int64(views),
				"youtube_comments": int64(comments),
			},
			Time:      time.Now(),
			Precision: "h",
		})
	}

	if val, ok := err.(*googleapi.Error); ok && val.Code == 403 {
		time.Sleep(time.Minute)
	}

	return err
}
