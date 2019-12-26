package queue

import (
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/queue/framework"
)

type GroupNewMessage struct {
	ID string `json:"id"`
}

func newGroupsHandler(messages []*framework.Message) {

	for _, message := range messages {

		payload := GroupNewMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			return
		}

		//
		err = ProduceGroup(GroupMessage{ID: payload.ID})
		if err != nil {
			log.Err(err)
		} else {
			message.Ack()
		}
	}
}
