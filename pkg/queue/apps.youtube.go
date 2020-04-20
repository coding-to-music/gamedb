package queue

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/helpers/influx"
	"github.com/gamedb/gamedb/pkg/log"
	influx "github.com/influxdata/influxdb1-client"
	"google.golang.org/api/googleapi"
)

type AppYoutubeMessage struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func appYoutubeHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		payload := AppYoutubeMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			continue
		}

		var wg sync.WaitGroup

		wg.Add(1)
		var youtubeViews uint64
		var youtubeComments uint64
		go func() {

			defer wg.Done()

			var err error
			youtubeViews, youtubeComments, err = appsYoutubeFetch(payload)
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

		// Save to Influx
		wg.Add(1)
		go func() {

			defer wg.Done()

			err = appsYoutubeSave(payload, youtubeViews, youtubeComments)
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

func appsYoutubeFetch(payload AppYoutubeMessage) (uint64, uint64, error) {

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

func appsYoutubeSave(payload AppYoutubeMessage, views uint64, comments uint64) (err error) {

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
