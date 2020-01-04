package queue

import (
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/log"
)

type TestMessage struct {
	ID int `json:"id"`
}

func testHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		log.Info(time.Now().String())

		message.Ack(false)
	}
}
