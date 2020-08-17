package queue

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/steam"
	"github.com/gamedb/gamedb/pkg/twitch"
	influx "github.com/influxdata/influxdb1-client"
	"github.com/nicklaw5/helix"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type AppPlayerMessage struct {
	IDs []int `json:"ids"`
}

func appPlayersHandler(message *rabbit.Message) {

	payload := AppPlayerMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		zap.S().Error(err, string(message.Message.Body))
		sendToFailQueue(message)
		return
	}

	// Get apps
	apps, err := mongo.GetAppsByID(payload.IDs, bson.M{"_id": 1, "twitch_id": 1, "player_peak_week": 1, "player_peak_alltime": 1})
	if err != nil {
		zap.S().Error(err, payload.IDs)
		sendToRetryQueue(message)
		return
	}

	for _, app := range apps {

		var wg sync.WaitGroup

		// Reads
		wg.Add(1)
		var twitchViewers int
		go func() {

			defer wg.Done()

			var err error
			twitchViewers, err = getAppTwitchStreamers(app.TwitchID)
			if err != nil {

				if strings.Contains(err.Error(), "read: connection reset by peer") ||
					strings.Contains(err.Error(), "i/o timeout") ||
					strings.Contains(err.Error(), "unexpected EOF") {
					zap.S().Info(err, payload.IDs)
				} else {
					zap.S().Error(err, payload.IDs)
				}

				sendToRetryQueue(message)
				return
			}
		}()

		wg.Add(1)
		var inGame int
		go func() {

			defer wg.Done()

			var err error
			inGame, err = getAppOnlinePlayers(app.ID)
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

			err = saveAppPlayerToInflux(app.ID, twitchViewers, inGame)
			if err != nil {
				zap.S().Error(err, payload.IDs)
				sendToRetryQueue(message)
				return
			}
		}()

		// Update app row
		if inGame > app.PlayerPeakAllTime {

			wg.Add(1)
			go func() {

				defer wg.Done()

				filter := bson.D{{"_id", app.ID}}
				update := bson.D{
					{"player_peak_alltime", inGame},
					{"player_peak_alltime_time", time.Now()},
				}

				_, err = mongo.UpdateOne(mongo.CollectionApps, filter, update)
				if err != nil {
					zap.S().Error(err, app.ID)
					sendToRetryQueue(message)
					return
				}

				// Clear cache
				err = memcache.Delete(memcache.MemcacheApp(app.ID).Key)
				if err != nil {
					zap.S().Error(err, app.ID)
					sendToRetryQueue(message)
					return
				}

				// Update in Elastic
				err = ProduceAppSearch(nil, app.ID)
				if err != nil {
					zap.S().Error(err, app.ID)
					sendToRetryQueue(message)
					return
				}
			}()
		}

		wg.Wait()

		if message.ActionTaken {
			continue
		}
	}

	//
	message.Ack()
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

	count, err = steam.GetSteamUnlimited().GetNumberOfCurrentPlayers(appID)
	err = steam.AllowSteamCodes(err, 404)
	return count, err
}

func saveAppPlayerToInflux(appID int, twitchViewers int, playersInGame int) (err error) {

	if twitchViewers == 0 && playersInGame == 0 {
		return nil
	}

	_, err = influxHelper.InfluxWrite(influxHelper.InfluxRetentionPolicyAllTime, influx.Point{
		Measurement: string(influxHelper.InfluxMeasurementApps),
		Tags: map[string]string{
			"app_id": strconv.Itoa(appID),
		},
		Fields: map[string]interface{}{
			"player_count":   playersInGame,
			"twitch_viewers": twitchViewers,
		},
		Time:      time.Now(),
		Precision: "m",
	})

	return err
}
