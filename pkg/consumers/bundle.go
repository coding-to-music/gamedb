package consumers

import (
	"github.com/gamedb/gamedb/pkg/consumers/framework"
	"github.com/gamedb/gamedb/pkg/log"
)

func bundleHandler(message framework.Message) {

	log.Info("bundle handler")

	err := message.SendToQueue(queues[framework.Producer][queueApps])
	log.Err(err)

}
