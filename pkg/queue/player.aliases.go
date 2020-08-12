package queue

import (
	"strconv"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/steam"
	"github.com/gamedb/gamedb/pkg/websockets"
	"go.mongodb.org/mongo-driver/bson"
)

type PlayersAliasesMessage struct {
	PlayerID int64 `json:"player_id"`
}

func (m PlayersAliasesMessage) Queue() rabbit.QueueName {
	return QueuePlayersAliases
}

func playerAliasesHandler(message *rabbit.Message) {

	payload := PlayersAliasesMessage{}

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
			Queue: "alias",
		}

		err = ProduceWebsocket(wsPayload, websockets.PagePlayer)
		if err != nil {
			log.Err(err, message.Message.Body)
		}
	}()

	aliases, err := steam.GetSteam().GetAliases(payload.PlayerID)
	if err == steamapi.ErrProfileMissing {
		message.Ack()
		return
	}
	err = steam.AllowSteamCodes(err)
	if err != nil {
		steam.LogSteamError(err, payload.PlayerID)
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
				log.Err(err, v.Time, payload.PlayerID)
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
		log.Err(err, payload.PlayerID)
		sendToRetryQueue(message)
		return
	}

	// Update player row
	update := bson.D{{"aliases", playerAliasStrings}}

	_, err = mongo.UpdateOne(mongo.CollectionPlayers, bson.D{{"_id", payload.PlayerID}}, update)
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

	message.Ack()
}
