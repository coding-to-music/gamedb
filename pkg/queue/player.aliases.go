package queue

import (
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/steam"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type PlayersAliasesMessage struct {
	PlayerID      int64 `json:"player_id"`
	PlayerRemoved bool  `json:"player_removed"`
}

func (m PlayersAliasesMessage) Queue() rabbit.QueueName {
	return QueuePlayersAliases
}

func playerAliasesHandler(message *rabbit.Message) {

	payload := PlayersAliasesMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToFailQueue(message)
		return
	}

	// Websocket
	defer sendPlayerWebsocket(payload.PlayerID, "alias", message)

	//
	if payload.PlayerRemoved {
		message.Ack()
		return
	}

	//
	aliases, b, err := steam.GetSteam().GetAliases(payload.PlayerID)
	if err == steamapi.ErrProfileMissing {
		message.Ack()
		return
	}
	err = steam.AllowSteamCodes(err)
	if err != nil {
		steam.LogSteamError(err, zap.Int64("player id", payload.PlayerID), zap.String("resp", string(b)))
		sendToRetryQueue(message)
		return
	}

	var playerAliases []mongo.PlayerAlias
	var playerAliasStrings []string

	for _, v := range aliases {

		var t time.Time

		t, err = time.Parse("2 Jan @ 3:04pm", v.Time)
		if err == nil {
			t = t.AddDate(time.Now().Year(), 0, 0)
		}
		if err != nil {

			t, err = time.Parse("2 Jan, 2006 @ 3:04pm", v.Time)
			if err != nil {
				log.ErrS(err, v.Time, payload.PlayerID)
				continue
			}
		}

		playerAliases = append(playerAliases, mongo.PlayerAlias{
			PlayerID:   payload.PlayerID,
			PlayerName: v.Alias,
			Time:       t.Unix(),
		})

		playerAliasStrings = append(playerAliasStrings, v.Alias)
	}

	err = mongo.ReplacePlayerAliases(playerAliases)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToRetryQueue(message)
		return
	}

	// Update player row
	update := bson.D{{"aliases", playerAliasStrings}}

	_, err = mongo.UpdateOne(mongo.CollectionPlayers, bson.D{{"_id", payload.PlayerID}}, update, nil)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToRetryQueue(message)
		return
	}

	// Clear player cache
	err = memcache.Delete(memcache.ItemPlayer(payload.PlayerID).Key)
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
