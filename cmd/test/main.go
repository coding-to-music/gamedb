package main

import (
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/queue"
)

func main() {

	config.SetVersion("test")
	log.Initialise([]log.LogName{log.LogNameTest})
	queue.Init(queue.AllProducerDefinitions)

}
