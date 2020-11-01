package queue

import (
	"sort"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/steam"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type PlayerGamesMessage struct {
	PlayerID                 int64     `json:"player_id"`
	PlayerCountry            string    `json:"player_country"`
	PlayerUpdated            time.Time `json:"player_updated"`
	SkipAchievements         bool      `json:"skip_achievements"`
	ForceAchievementsRefresh bool      `json:"force_achievements_refresh"`
	OldAchievementCount      int       `json:"old_achievement_count"`
	OldAchievementCount100   int       `json:"old_achievement_count_100"`
	OldAchievementCountApps  int       `json:"old_achievement_count_apps"`
}

func (m PlayerGamesMessage) Queue() rabbit.QueueName {
	return QueuePlayersGames
}

// Always updates player_app rows as the playtime will change
func playerGamesHandler(message *rabbit.Message) {

	payload := PlayerGamesMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToFailQueue(message)
		return
	}

	// Websocket
	defer sendPlayerWebsocket(payload.PlayerID, "game", message)

	//
	if message.Attempt() > 10 {
		message.Ack()
		return
	}

	//
	updatePlayer := bson.D{}

	// Grab games from Steam
	resp, err := steam.GetSteam().GetOwnedGames(payload.PlayerID)
	err = steam.AllowSteamCodes(err)
	if err != nil {
		steam.LogSteamError(err, string(message.Message.Body))
		sendToRetryQueue(message)
		return
	}

	// Save count
	updatePlayer = append(updatePlayer, bson.E{Key: "games_count", Value: len(resp.Games)})

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

	updatePlayer = append(updatePlayer,
		bson.E{Key: "play_time", Value: playtime},
		bson.E{Key: "play_time_windows", Value: playtimeWindows},
		bson.E{Key: "play_time_mac", Value: playtimeMac},
		bson.E{Key: "play_time_linux", Value: playtimeLinux},
	)

	// Getting missing price info from Mongo
	gameRows, err := mongo.GetAppsByID(appIDs, bson.M{"_id": 1, "prices": 1, "type": 1})
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
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
		playerApps[gameRow.ID].AppPriceHour = appPriceHour[gameRow.ID]
	}

	updatePlayer = append(updatePlayer, bson.E{Key: "games_by_type", Value: gamesByType})

	// Save playerApps to Mongo
	err = mongo.UpdatePlayerApps(playerApps)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToRetryQueue(message)
		return
	}

	// Get top game for background
	if len(appIDs) > 0 {

		sort.Slice(appIDs, func(i, j int) bool {
			return playerApps[appIDs[i]].AppTime > playerApps[appIDs[j]].AppTime
		})

		updatePlayer = append(updatePlayer, bson.E{Key: "background_app_id", Value: appIDs[0]})
	}

	// Save stats to player
	var gameStats = mongo.PlayerAppStatsTemplate{}
	for _, v := range playerApps {

		gameStats.All.AddApp(v.AppTime, appPrices[v.AppID], appPriceHour[v.AppID])
		if v.AppTime > 0 {
			gameStats.Played.AddApp(v.AppTime, appPrices[v.AppID], appPriceHour[v.AppID])
		}
	}

	updatePlayer = append(updatePlayer, bson.E{Key: "game_stats", Value: gameStats})

	if !payload.SkipAchievements || payload.ForceAchievementsRefresh {
		if payload.PlayerUpdated.Before(time.Now().Add(time.Hour * 24 * 13 * -1)) { // Just under 2 weeks
			for _, v := range resp.Games {
				if v.PlaytimeForever > 0 {
					err = ProducePlayerAchievements(
						payload.PlayerID, v.AppID, payload.ForceAchievementsRefresh,
						payload.OldAchievementCount, payload.OldAchievementCount100, payload.OldAchievementCountApps,
					)
					if err != nil {
						log.ErrS(err)
					}
				}
			}
			err = ProducePlayerAchievements(
				payload.PlayerID, 0, false,
				payload.OldAchievementCount, payload.OldAchievementCount100, payload.OldAchievementCountApps,
			)
			if err != nil {
				log.ErrS(err)
			}
		}
	}

	// Update player row
	_, err = mongo.UpdateOne(mongo.CollectionPlayers, bson.D{{"_id", payload.PlayerID}}, updatePlayer)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToRetryQueue(message)
		return
	}

	// Save to Influx
	err = savePlayerStatsToInflux(payload.PlayerID, map[string]interface{}{
		influx.InfPlayersGames.String():    len(resp.Games),
		influx.InfPlayersPlaytime.String(): playtime,
	})
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToRetryQueue(message)
		return
	}

	// Clear player cache
	err = memcache.Delete(memcache.MemcachePlayer(payload.PlayerID).Key)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToRetryQueue(message)
		return
	}

	// Update Elastic
	err = ProducePlayerSearch(nil, payload.PlayerID)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToRetryQueue(message)
		return
	}

	//
	message.Ack()
}
