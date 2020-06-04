package queue

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/helpers/influx"
	"github.com/gamedb/gamedb/pkg/helpers/steam"
	"github.com/gamedb/gamedb/pkg/helpers/twitch"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	influx "github.com/influxdata/influxdb1-client"
	"github.com/nicklaw5/helix"
	"go.mongodb.org/mongo-driver/bson"
)

type AppPlayerMessage struct {
	IDs []int `json:"ids"`
}

func appPlayersHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		payload := AppPlayerMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			continue
		}

		// Get apps
		apps, err := mongo.GetAppsByID(payload.IDs, bson.M{"_id": 1, "twitch_id": 1})
		if err != nil {
			log.Err(err, payload.IDs)
			sendToRetryQueue(message)
			continue
		}

		for _, app := range apps {

			var wg sync.WaitGroup

			// Reads
			wg.Add(1)
			var viewers int
			go func() {

				defer wg.Done()

				var err error
				viewers, err = getAppTwitchStreamers(app.TwitchID)
				if err != nil {

					if strings.Contains(err.Error(), "read: connection reset by peer") ||
						strings.Contains(err.Error(), "unexpected EOF") {
						log.Info(err, payload.IDs)
					} else {
						log.Err(err, payload.IDs)
					}

					sendToRetryQueue(message)
					return
				}
			}()

			wg.Add(1)
			var appPlayersNow int
			go func() {

				defer wg.Done()

				var err error
				appPlayersNow, err = getAppOnlinePlayers(app.ID)
				if err != nil {
					steam.LogSteamError(err, payload.IDs)
					sendToRetryQueue(message)
					return
				}
			}()

			wg.Wait()

			if message.ActionTaken {
				continue
			}

			// Save counts to Influx
			wg.Add(1)
			go func() {

				defer wg.Done()

				err = saveAppPlayerToInflux(app.ID, viewers, appPlayersNow)
				if err != nil {
					log.Err(err, payload.IDs)
					sendToRetryQueue(message)
					return
				}
			}()

			wg.Wait()

			if message.ActionTaken {
				continue
			}
		}

		//
		message.Ack(false)
	}
}

func getAppTwitchStreamers(twitchID int) (viewers int, err error) {

	if twitchID > 0 {

		client, err := twitch.GetTwitch()
		if err != nil {
			return 0, err
		}

		resp, err := client.GetStreams(&helix.StreamsParams{First: 100, GameIDs: []string{strconv.Itoa(twitchID)}, Language: []string{"en"}})
		if err != nil {
			return 0, err
		}

		for _, v := range resp.Data.Streams {
			viewers += v.ViewerCount
		}
	}

	return viewers, nil
}

func getAppOnlinePlayers(appID int) (count int, err error) {

	count, _, err = steam.GetSteamUnlimited().GetNumberOfCurrentPlayers(appID)
	err = steam.AllowSteamCodes(err, 404)
	return count, err
}

func saveAppPlayerToInflux(appID int, viewers int, players int) (err error) {

	_, err = influxHelper.InfluxWrite(influxHelper.InfluxRetentionPolicyAllTime, influx.Point{
		Measurement: string(influxHelper.InfluxMeasurementApps),
		Tags: map[string]string{
			"app_id": strconv.Itoa(appID),
		},
		Fields: map[string]interface{}{
			"player_count":   players,
			"twitch_viewers": viewers,
		},
		Time:      time.Now(),
		Precision: "m",
	})

	return err
}
