package consumers

import (
	"time"

	"github.com/gamedb/gamedb/pkg/consumers/framework"
	"github.com/gamedb/gamedb/pkg/log"
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
