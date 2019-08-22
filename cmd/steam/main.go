package main

import (
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/queue"
)

func main() {

	// Load consumer
	log.Info("Starting Steam consumers")

	q := queue.QueueRegister[queue.QueueSteam]
	q.SteamClient = steamClient

	go q.ConsumeMessages()

	// Load Steam
	log.Info("Starting Steam client")
	InitSteam()

	//
	helpers.KeepAlive()
}
