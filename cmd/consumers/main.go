package main

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/sql"
)

var version string

func main() {

	config.SetVersion(version)
	log.Initialise([]log.LogName{log.LogNameConsumers})

	// Get API key
	err := sql.GetAPIKey("consumer")
	if err != nil {
		log.Critical(err)
		return
	}

	// Load pubsub
	log.Info("Listening to PubSub for memcache")
	go memcache.ListenToPubSubMemcache()

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
	if config.IsLocal() {
		log.Info("Starting consumers profiling")
		go func() {
			err := http.ListenAndServe("localhost:6060", nil)
			log.Critical(err)
		}()
	}

	// Load consumers
	queue.Init(queue.QueueDefinitions, true)

	helpers.KeepAlive()
}
