package queue

import (
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/steam"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
)

type PlayersAliasesMessage struct {
	PlayerID int64 `json:"player_id"`
}

func playerAliasesHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		payload := PlayersAliasesMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			continue
		}

		aliases, _, err := steam.GetSteam().GetAliases(payload.PlayerID)
		err = steam.AllowSteamCodes(err)
		if err != nil {
			steam.LogSteamError(err, payload.PlayerID)
			sendToRetryQueue(message)
			continue
		}

		var playerAliases []mongo.PlayerAlias

		for _, v := range aliases {

			var t time.Time

			t, err = time.Parse("2 Jan @ 3:04pm", v.Time)
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
		}

		err = mongo.UpdatePlayerAliases(playerAliases)
		if err != nil {
			log.Err(err, payload.PlayerID)
			sendToRetryQueue(message)
			continue
		}

		message.Ack(false)
	}
}
