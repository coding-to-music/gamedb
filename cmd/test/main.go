package main

import (
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/search"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/queue"
)

func main() {

	config.Init("test", helpers.GetIP())
	log.Initialise([]log.LogName{log.LogNameTest})
	queue.Init(queue.AllProducerDefinitions)

	err := queue.ProduceSearch(queue.SearchMessage{
		ID:      578080,
		Name:    "PLAYERUNKNOWN'S BATTLEGROUNDS",
		Aliases: []string{"pubg"},
		Type:    search.SearchTypeApp,
	})

	log.Err(err)

	log.Info("Queued")
}
