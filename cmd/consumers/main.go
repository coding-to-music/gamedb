package main

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/queue"
)

var version string
var commits string

func main() {

	err := config.Init(version, commits, helpers.GetIP())
	log.InitZap(log.LogNameConsumers)
	if err != nil {
		log.FatalS(err)
		return
	}

	// Get API key
	err = mysql.GetConsumer("consumer")
	if err != nil {
		log.FatalS(err)
		return
	}

	// Load Discord
	// discord, err := discordgo.New("Bot " + config.C.DiscordChangesBotToken)
	// if err != nil {
	// 	log.FatalS(err)
	// 	return
	// }
	//
	// // Not used right now
	// err = discord.Open()
	// if err != nil {
	// 	log.FatalS(err)
	// 	return
	// }
	//
	// queue.SetDiscordClient(discord)

	// Profiling
	if !config.IsConsumer() {
		log.Info("Starting consumers profiling")
		go func() {
			err := http.ListenAndServe(":6062", nil)
			if err != nil {
				log.FatalS(err)
			}
		}()
	}

	// Load consumers
	queue.Init(queue.ConsumersDefinitions)

	helpers.KeepAlive()
}
