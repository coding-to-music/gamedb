package queue

import (
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/helpers/influx"
	"github.com/gamedb/gamedb/pkg/log"
	influx "github.com/influxdata/influxdb1-client"
	"google.golang.org/api/googleapi"
)

type AppDailyMessage struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func appDailyHandler(messages []*rabbit.Message) {

	var points []influx.Point
	var lastMessage *rabbit.Message

	for _, message := range messages {

		payload := AppDailyMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			return
		}

		name := strings.TrimSpace(payload.Name)

		if name != "" {

			views, comments, err := getYouTubeStats(name)
			if err != nil {

				if val, ok := err.(*googleapi.Error); ok && val.Code == 403 {
					time.Sleep(time.Minute)
				}

				log.Err(err, payload.ID)
				sendToRetryQueue(messages...)
				return
			}

			if views > 0 || comments > 0 {

				points = append(points, influx.Point{
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
		}

		//
		if message.PercentOfBatch() == 100 {
			lastMessage = message
		}
	}

	_, err := influxHelper.InfluxWriteMany(influxHelper.InfluxRetentionPolicyAllTime, influx.BatchPoints{
		Points:          points,
		Database:        influxHelper.InfluxGameDB,
		RetentionPolicy: influxHelper.InfluxRetentionPolicyAllTime.String(),
		Precision:       "h",
	})

	if err != nil {
		log.Err(err)
		sendToRetryQueue(messages...)
		return
	}

	if lastMessage != nil {
		lastMessage.Ack(true)
	}
}

func getYouTubeStats(name string) (uint64, uint64, error) {

	// `part` can be:
	// id, snippet, contentDetails, fileDetails, player, processingDetails, recordingDetails, statistics, status, suggestions, topicDetails

	// Get video IDs from search
	searchResponse, err := helpers.YoutubeService.Search.List("id").
		Context(helpers.YoutubeContext).
		MaxResults(50).
		SafeSearch("none").
		Type("video").
		Q(name).
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
