package queue

import (
	"github.com/Philipp15b/go-steam"
	"github.com/gamedb/gamedb/pkg/consumers"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/streadway/amqp"
)

type steamMessage struct {
	baseMessage
	Message steamMessageInner `json:"message"`
}

type steamMessageInner struct {
	AppIDs     []int   `json:"app_ids"`
	PackageIDs []int   `json:"package_ids"`
	PlayerIDs  []int64 `json:"player_ids"`
}

type steamQueue struct {
	SteamClient *steam.Client
}

func (q steamQueue) processMessages(msgs []amqp.Delivery) {

	message := steamMessage{}

	err := helpers.Unmarshal(msgs[0].Body, &message)
	if err != nil {
		log.Err(err, msgs[0].Body)
		ackFail(msgs[0], &message)
		return
	}

	//
	for _, v := range message.Message.PlayerIDs {

		err = consumers.ProducePlayer(consumers.PlayerMessage{ID: v})
		if err != nil {
			log.Err(err, msgs[0].Body)
		}
	}

	//
	payload := consumers.SteamMessage{}
	payload.AppIDs = message.Message.AppIDs
	payload.PackageIDs = message.Message.PackageIDs

	err = consumers.ProduceSteam(payload)
	if err != nil {
		log.Err(err, msgs[0].Body)
		ackRetry(msgs[0], &message)
	} else {
		message.ack(msgs[0])
	}
}
