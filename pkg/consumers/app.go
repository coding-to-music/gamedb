package consumers

import (
	"github.com/gamedb/gamedb/pkg/consumers/framework"
	"github.com/gamedb/gamedb/pkg/log"
)

func appHandler(message framework.Message) {

	log.Info("app handler")

	err := message.SendToQueue(queues[framework.Producer][queueBundles])
	log.Err(err)

}
