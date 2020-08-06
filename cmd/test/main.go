package main

import (
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/queue"
)

func main() {

	config.Init("test", "0", helpers.GetIP())
	log.Initialise(log.LogNameTest)
	queue.Init(queue.AllProducerDefinitions)

}
