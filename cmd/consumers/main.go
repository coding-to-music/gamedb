package main

import (
	"github.com/gamedb/website/pkg/config"
	"github.com/gamedb/website/pkg/helpers"
	"github.com/gamedb/website/pkg/log"
	"github.com/gamedb/website/pkg/queue"
)

func main() {

	log.Info("Starting consumers")

	if config.Config.EnableConsumers.GetBool() {
		go func() {
			log.Info("Starting consumers")
			queue.RunConsumers()
		}()
	}

	helpers.KeepAlive()
}
