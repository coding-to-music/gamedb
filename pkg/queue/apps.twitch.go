package queue

import (
	"strconv"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/twitch"
	"github.com/nicklaw5/helix"
	"go.mongodb.org/mongo-driver/bson"
)

type AppTwitchMessage struct {
	ID int `json:"id"`
}

func (m AppTwitchMessage) Queue() rabbit.QueueName {
	return QueueAppsTwitch
}

func appTwitchHandler(message *rabbit.Message) {

	payload := AppTwitchMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err, message.Message.Body)
		sendToFailQueue(message)
		return
	}

	app, err := mongo.GetApp(payload.ID, false)
	if err != nil {
		log.Err(err, payload.ID)
		sendToRetryQueue(message)
		return
	}

	if app.Name != "" && app.Type != "game" && (app.TwitchID == 0 || app.TwitchURL == "") {

		client, err := twitch.GetTwitch()
		if err != nil {
			log.Err(err, payload.ID)
			sendToRetryQueue(message)
			return
		}

		resp, err := client.GetGames(&helix.GamesParams{Names: []string{app.Name}})
		if err != nil {
			log.Err(err, payload.ID)
			sendToRetryQueue(message)
			return
		}

		if len(resp.Data.Games) > 0 {

			i, err := strconv.Atoi(resp.Data.Games[0].ID)
			if err != nil {
				log.Err(err, payload.ID)
				sendToRetryQueue(message)
				return
			}

			var update = bson.D{
				{"twitch_id", i},
				{"twitch_url", resp.Data.Games[0].Name},
			}

			_, err = mongo.UpdateOne(mongo.CollectionApps, bson.D{{"_id", payload.ID}}, update)
			if err != nil {
				log.Err(err, payload.ID)
				sendToRetryQueue(message)
				return
			}

			err = memcache.Delete(memcache.MemcacheApp(payload.ID).Key)
			if err != nil {
				log.Err(err, payload.ID)
				sendToRetryQueue(message)
				return
			}
		}
	}

	message.Ack(false)
}
