package queue

import (
	"sort"
	"strconv"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/steam"
	"github.com/gamedb/gamedb/pkg/websockets"
	influx "github.com/influxdata/influxdb1-client"
	"go.mongodb.org/mongo-driver/bson"
)

type PlayerGamesMessage struct {
	PlayerID                 int64     `json:"player_id"`
	PlayerCountry            string    `json:"player_country"`
	PlayerUpdated            time.Time `json:"player_updated"`
	SkipAchievements         bool      `json:"skip_achievements"`
	ForceAchievementsRefresh bool      `json:"force_achievements_refresh"`
}

func (m PlayerGamesMessage) Queue() rabbit.QueueName {
	return QueuePlayersGames
}

// Always updates player_app rows as the playtime will change
func playerGamesHandler(message *rabbit.Message) {

	payload := PlayerGamesMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err, message.Message.Body)
		sendToFailQueue(message)
		return
	}

	// Websocket
	defer func() {

		wsPayload := PlayerPayload{
			ID:    strconv.FormatInt(payload.PlayerID, 10),
			Queue: "game",
		}

		err = ProduceWebsocket(wsPayload, websockets.PagePlayer)
		if err != nil {
			log.Err(err, message.Message.Body)
		}
	}()

	//
	update := bson.D{}

	// Grab games from Steam
	resp, err := steam.GetSteam().GetOwnedGames(payload.PlayerID)
	err = steam.AllowSteamCodes(err)
	if err != nil {
		steam.LogSteamError(err, message.Message.Body)
		sendToRetryQueue(message)
		return
	}

	// Save count
	update = append(update, bson.E{Key: "games_count", Value: len(resp.Games)})

	// Start creating PlayerApp's
	var playerApps = map[int]*mongo.PlayerApp{}
	var appPrices = map[int]map[string]int{}
	var appPriceHour = map[int]map[string]float64{}
	var appIDs []int

	var playtime = 0
	var playtimeWindows = 0
	var playtimeMac = 0
	var playtimeLinux = 0

	for _, v := range resp.Games {

		playtime += v.PlaytimeForever
		playtimeWindows += v.PlaytimeWindows
		playtimeMac += v.PlaytimeMac
		playtimeLinux += v.PlaytimeLinux

		appIDs = append(appIDs, v.AppID)
		playerApps[v.AppID] = &mongo.PlayerApp{
			PlayerID:      payload.PlayerID,
			PlayerCountry: payload.PlayerCountry,
			AppID:         v.AppID,
			AppName:       v.Name,
			AppIcon:       v.ImgIconURL,
			AppTime:       v.PlaytimeForever,
		}
		appPrices[v.AppID] = map[string]int{}
		appPriceHour[v.AppID] = map[string]float64{}
	}

	update = append(update,
		bson.E{Key: "play_time", Value: playtime},
		bson.E{Key: "play_time_windows", Value: playtimeWindows},
		bson.E{Key: "play_time_mac", Value: playtimeMac},
		bson.E{Key: "play_time_linux", Value: playtimeLinux},
	)

	// Getting missing price info from MySQL
	gameRows, err := mongo.GetAppsByID(appIDs, bson.M{"_id": 1, "prices": 1, "type": 1})
	if err != nil {
		log.Err(err, message.Message.Body)
		sendToRetryQueue(message)
		return
	}

	gamesByType := map[string]int{}

	for _, gameRow := range gameRows {

		// Set games by type
		if _, ok := gamesByType[gameRow.GetType()]; ok {
			gamesByType[gameRow.GetType()]++
		} else {
			gamesByType[gameRow.GetType()] = 1
		}

		//
		for code, vv := range gameRow.Prices {

			vv = gameRow.Prices.Get(code)

			appPrices[gameRow.ID][string(code)] = vv.Final
			if appPrices[gameRow.ID][string(code)] > 0 && playerApps[gameRow.ID].AppTime == 0 {
				appPriceHour[gameRow.ID][string(code)] = -1 // Infinite
			} else if appPrices[gameRow.ID][string(code)] > 0 && playerApps[gameRow.ID].AppTime > 0 {
				appPriceHour[gameRow.ID][string(code)] = (float64(appPrices[gameRow.ID][string(code)]) / 100) / (float64(playerApps[gameRow.ID].AppTime) / 60) * 100
			} else {
				appPriceHour[gameRow.ID][string(code)] = 0 // Free
			}
		}

		//
		playerApps[gameRow.ID].AppPrices = appPrices[gameRow.ID]
		log.Err(err)

		//
		playerApps[gameRow.ID].AppPriceHour = appPriceHour[gameRow.ID]
		log.Err(err)
	}

	update = append(update, bson.E{Key: "games_by_type", Value: gamesByType})

	// Save playerApps to Mongo
	err = mongo.UpdatePlayerApps(playerApps)
	if err != nil {
		log.Err(err, message.Message.Body)
		sendToRetryQueue(message)
		return
	}

	// Get top game for background
	if len(appIDs) > 0 {

		sort.Slice(appIDs, func(i, j int) bool {
			return playerApps[appIDs[i]].AppTime > playerApps[appIDs[j]].AppTime
		})

		update = append(update, bson.E{Key: "background_app_id", Value: appIDs[0]})
	}

	// Save stats to player
	var gameStats = mongo.PlayerAppStatsTemplate{}
	for _, v := range playerApps {

		gameStats.All.AddApp(v.AppTime, appPrices[v.AppID], appPriceHour[v.AppID])
		if v.AppTime > 0 {
			gameStats.Played.AddApp(v.AppTime, appPrices[v.AppID], appPriceHour[v.AppID])
		}
	}

	update = append(update, bson.E{Key: "game_stats", Value: gameStats})

	if !payload.SkipAchievements || payload.ForceAchievementsRefresh {
		if payload.PlayerUpdated.Before(time.Now().Add(time.Hour * 24 * 13 * -1)) { // Just under 2 weeks
			for _, v := range resp.Games {
				if v.PlaytimeForever > 0 {
					err = ProducePlayerAchievements(payload.PlayerID, v.AppID, payload.ForceAchievementsRefresh)
					log.Err(err)
				}
			}
			err = ProducePlayerAchievements(payload.PlayerID, 0, false)
			log.Err(err)
		}
	}

	// Update player row
	_, err = mongo.UpdateOne(mongo.CollectionPlayers, bson.D{{"_id", payload.PlayerID}}, update)
	if err != nil {
		log.Err(err, payload.PlayerID)
		sendToRetryQueue(message)
		return
	}

	// Save to Influx
	point := influx.Point{
		Measurement: string(influxHelper.InfluxMeasurementPlayers),
		Tags: map[string]string{
			"player_id": strconv.FormatInt(payload.PlayerID, 10),
		},
		Fields: map[string]interface{}{
			"games":    len(resp.Games),
			"playtime": playtime,
		},
		Time:      time.Now(),
		Precision: "m",
	}

	_, err = influxHelper.InfluxWrite(influxHelper.InfluxRetentionPolicyAllTime, point)
	if err != nil {
		log.Err(err, payload.PlayerID)
		sendToRetryQueue(message)
		return
	}

	// Clear player cache
	err = memcache.Delete(memcache.MemcachePlayer(payload.PlayerID).Key)
	if err != nil {
		log.Err(err, payload.PlayerID)
		sendToRetryQueue(message)
		return
	}

	message.Ack(false)
}
