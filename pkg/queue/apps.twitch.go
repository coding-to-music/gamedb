package queue

import (
	"strconv"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue/helpers/twitch"
	"github.com/nicklaw5/helix"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type AppTwitchMessage struct {
	AppID int `json:"id"`
}

func (m AppTwitchMessage) Queue() rabbit.QueueName {
	return QueueAppsTwitch
}

func appTwitchHandler(message *rabbit.Message) {

	payload := AppTwitchMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToFailQueue(message)
		return
	}

	app, err := mongo.GetApp(payload.AppID)
	if err != nil {
		log.ErrS(err, payload.AppID)
		sendToRetryQueue(message)
		return
	}

	if (app.TwitchID > 0 && app.TwitchURL != "") || app.Name == "" {
		message.Ack()
		return
	}

	client, err := twitch.GetTwitch()
	if err != nil {
		log.ErrS(err, payload.AppID)
		sendToRetryQueue(message)
		return
	}

	resp, err := client.GetGames(&helix.GamesParams{Names: []string{app.Name}})
	if err != nil {
		log.ErrS(err, payload.AppID)
		sendToRetryQueue(message)
		return
	}

	if len(resp.Data.Games) == 0 {
		message.Ack()
		return
	}

	i, err := strconv.Atoi(resp.Data.Games[0].ID)
	if err != nil {
		log.ErrS(err, payload.AppID)
		sendToRetryQueue(message)
		return
	}

	var update = bson.D{
		{"twitch_id", i},
		{"twitch_url", resp.Data.Games[0].Name},
	}

	_, err = mongo.UpdateOne(mongo.CollectionApps, bson.D{{"_id", payload.AppID}}, update)
	if err != nil {
		log.ErrS(err, payload.AppID)
		sendToRetryQueue(message)
		return
	}

	err = memcache.Client().Delete(memcache.ItemApp(payload.AppID).Key)
	if err != nil {
		log.ErrS(err, payload.AppID)
		sendToRetryQueue(message)
		return
	}

	message.Ack()
}
