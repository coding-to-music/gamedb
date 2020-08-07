package main

import (
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/queue"
)

var version string
var commits string

func main() {

	config.Init(version, commits, helpers.GetIP())
	log.Initialise(log.LogNameTest)
	queue.Init(queue.AllProducerDefinitions)

	//

	helpers.KeepAlive()
}
