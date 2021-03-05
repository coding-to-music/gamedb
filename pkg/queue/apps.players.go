package queue

import (
	"strconv"
	"sync"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/cenkalti/backoff/v4"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue/helpers/twitch"
	"github.com/gamedb/gamedb/pkg/steam"
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
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToFailQueue(message)
		return
	}

	// Also push to influx queue
	err = ProduceAppsInflux(payload.IDs)
	if err != nil {
		log.ErrS(err, payload.IDs)
		sendToRetryQueue(message)
		return
	}

	// Get apps
	apps, err := mongo.GetAppsByID(payload.IDs, bson.M{"_id": 1, "twitch_id": 1, "player_peak_week": 1, "player_peak_alltime": 1})
	if err != nil {
		log.ErrS(err, payload.IDs)
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
				log.Err("Getting twitch streams", zap.Error(err), zap.Ints("ids", payload.IDs))
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
				steam.LogSteamError(err, zap.Ints("app ids", payload.IDs))
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
				log.ErrS(err, payload.IDs)
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
					log.ErrS(err, app.ID)
					sendToRetryQueue(message)
					return
				}

				// Clear cache
				err = memcache.Client().Delete(memcache.ItemApp(app.ID).Key)
				if err != nil {
					log.ErrS(err, app.ID)
					sendToRetryQueue(message)
					return
				}

				// No need to update in Elastic
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

		var resp *helix.StreamsResponse

		// Retry call
		operation := func() (err error) {

			client, err := twitch.GetTwitch()
			if err != nil {
				return err
			}

			resp, err = client.GetStreams(&helix.StreamsParams{First: 100, GameIDs: []string{strconv.Itoa(twitchID)}, Language: []string{"en"}})
			return err
		}

		notify := func(err error, t time.Duration) {
			log.Info("Getting twitch streams", zap.Int("twitch id", twitchID), zap.Error(err))
		}

		policy := backoff.NewExponentialBackOff()
		policy.InitialInterval = time.Second * 2

		err = backoff.RetryNotify(operation, backoff.WithMaxRetries(policy, 5), notify)
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

	// Used to use unlimited Steam
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
