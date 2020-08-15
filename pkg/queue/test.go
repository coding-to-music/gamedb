package queue

import (
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"go.uber.org/zap"
)

type TestMessage struct {
	ID int `json:"id"`
}

func testHandler(message *rabbit.Message) {

	payload := TestMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		zap.S().Error(err, message.Message.Body)
		sendToFailQueue(message)
		return
	}

	zap.S().Info(payload.ID, time.Now().String())

	message.Ack()
}
