package main

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/queue"
)

var version string

func main() {

	config.Init(version, helpers.GetIP())
	log.Initialise(log.LogNameConsumers)

	// Get API key
	err := mysql.GetAPIKey("consumer")
	if err != nil {
		log.Critical(err)
		return
	}

	// Load Discord
	discord, err := discordgo.New("Bot " + config.Config.DiscordChangesBotToken.Get())
	if err != nil {
		panic(err)
	}

	// Not used right now
	// err = discord.Open()
	// if err != nil {
	// 	panic(err)
	// }

	queue.SetDiscordClient(discord)

	// Load PPROF
	if !config.IsConsumer() {
		log.Info("Starting consumers profiling")
		go func() {
			err := http.ListenAndServe("localhost:6061", nil)
			log.Critical(err)
		}()
	}

	// Load consumers
	queue.Init(queue.ConsumersDefinitions)

	helpers.KeepAlive()
}
