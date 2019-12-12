package queue

import (
	"github.com/gamedb/gamedb/pkg/consumers"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/streadway/amqp"
)

type playerMessage struct {
	baseMessage
	Message playerMessageInner `json:"message"`
}

type playerMessageInner struct {
	ID            int64  `json:"id"`
	Eresult       int32  `json:"eresult,omitempty"`
	SteamidFriend int64  `json:"steamid_friend,omitempty"`
	TimeCreated   uint32 `json:"time_created,omitempty"`
	RealName      string `json:"real_name,omitempty"`
	CityName      string `json:"city_name,omitempty"`
	StateName     string `json:"state_name,omitempty"`
	CountryName   string `json:"country_name,omitempty"`
	Headline      string `json:"headline,omitempty"`
	Summary       string `json:"summary,omitempty"`
}

type playerQueue struct {
}

func (q playerQueue) processMessages(msgs []amqp.Delivery) {

	message := playerMessage{}

	err := helpers.Unmarshal(msgs[0].Body, &message)
	if err != nil {
		log.Err(err, msgs[0].Body)
		ackFail(msgs[0], &message)
		return
	}

	err = consumers.ProducePlayer(message.Message.ID)
	if err != nil {
		log.Err(err, msgs[0].Body)
		ackRetry(msgs[0], &message)
	} else {
		message.ack(msgs[0])
	}
}
