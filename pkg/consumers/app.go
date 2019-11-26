package consumers

import (
	"fmt"

	"github.com/gamedb/gamedb/pkg/consumers/framework"
	"github.com/gamedb/gamedb/pkg/log"
)

func appHandler(message framework.Message) {

	log.Info("app handler")

	x:= queues
	fmt.Println(x)

	err := message.SendToQueue(queues[CProducer][queueBundles])
	log.Err(err)
}
