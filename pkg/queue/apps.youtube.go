package queue

import (
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/queue/helpers/youtube"
	influx "github.com/influxdata/influxdb1-client"
	"go.uber.org/zap"
	"google.golang.org/api/googleapi"
)

var youtubeOverLimitAt time.Time

type AppYoutubeMessage struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func appYoutubeHandler(message *rabbit.Message) {

	if time.Now().Sub(youtubeOverLimitAt) < time.Hour {
		message.Ack()
		return
	}

	payload := AppYoutubeMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToFailQueue(message)
		return
	}

	if config.C.YoutubeAPIKey == "" {
		log.Err("Missing environment variables")
		sendToFailQueue(message)
		return
	}

	payload.Name = strings.TrimSpace(payload.Name)

	if payload.Name == "" {
		message.Ack()
		return
	}

	client, ctx, err := youtube.GetYouTube()
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToRetryQueue(message)
		return
	}

	// `part` can be:
	// id, snippet, contentDetails, fileDetails, player, processingDetails, recordingDetails, statistics, status, suggestions, topicDetails

	// Get video IDs from search
	searchRequest := client.Search.List([]string{"id"}).
		Context(ctx).
		MaxResults(50).
		SafeSearch("none").
		Type("video").
		Q(payload.Name).
		Order("viewCount").
		PublishedAfter(time.Now().Add(-time.Hour * 24 * 7).Format(time.RFC3339))

	searchResponse, err := searchRequest.Do()
	if err != nil {

		if val, ok := err.(*googleapi.Error); ok {
			if val.Code == 403 {
				youtubeOverLimitAt = time.Now()
			}
		}

		log.ErrS(err, payload.ID)
		sendToRetryQueue(message)
		return
	}

	var ids []string
	for _, v := range searchResponse.Items {
		ids = append(ids, v.Id.VideoId)
	}

	// Get video statistics from IDs
	listRequest := client.Videos.
		List([]string{"statistics"}).
		Id(strings.Join(ids, ","))

	listResponse, err := listRequest.Do()
	if err != nil {

		if val, ok := err.(*googleapi.Error); ok {
			if val.Code == 403 {
				youtubeOverLimitAt = time.Now()
			}
		}

		log.ErrS(err, payload.ID)
		sendToRetryQueue(message)
		return
	}

	var views uint64
	var comments uint64
	for _, v := range listResponse.Items {
		views += v.Statistics.ViewCount
		comments += v.Statistics.CommentCount
	}

	// Save to Influx
	if views > 0 || comments > 0 {

		point := influx.Point{
			Measurement: string(influxHelper.InfluxMeasurementApps),
			Tags: map[string]string{
				"app_id": strconv.Itoa(payload.ID),
			},
			Fields: map[string]interface{}{
				"youtube_views":    int64(views),
				"youtube_comments": int64(comments),
				"youtube_videos":   searchResponse.PageInfo.TotalResults,
			},
			Time:      time.Now(),
			Precision: "h",
		}

		_, err = influxHelper.InfluxWrite(influxHelper.InfluxRetentionPolicyAllTime, point)
		if err != nil {
			log.ErrS(err, payload.ID)
			if val, ok := err.(*googleapi.Error); ok && val.Code == 403 {
				time.Sleep(time.Minute)
			}
			sendToRetryQueue(message)
			return
		}
	}

	message.Ack()
}
