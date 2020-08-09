package queue

import (
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
)

type TestMessage struct {
	ID int `json:"id"`
}

func testHandler(message *rabbit.Message) {

	payload := TestMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err, message.Message.Body)
		sendToFailQueue(message)
		return
	}

	log.Info(payload.ID, time.Now().String())

	message.Ack()
}
