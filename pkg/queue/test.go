package queue

import (
	"time"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/queue/framework"
)

type TestMessage struct {
	ID int `json:"id"`
}

func testHandler(messages []*framework.Message) {

	for _, message := range messages {

		log.Info(time.Now().String())

		message.Ack()

	}
}
