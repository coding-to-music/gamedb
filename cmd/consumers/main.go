package main

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/queue"
)

func main() {

	err := config.Init(helpers.GetIP())
	log.InitZap(log.LogNameConsumers)
	defer log.Flush()
	if err != nil {
		log.ErrS(err)
		return
	}

	//
	log.Info("Starting consumers")

	// Get API key
	err = mysql.GetConsumer("consumer")
	if err != nil {
		log.ErrS(err)
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
	if config.IsLocal() {
		go func() {
			err := http.ListenAndServe(":6062", nil)
			if err != nil {
				log.ErrS(err)
			}
		}()
	}

	// Load consumers
	queue.Init(queue.ConsumersDefinitions)

	helpers.KeepAlive(func() {
		influx.GetWriter().Flush()
	})
}
