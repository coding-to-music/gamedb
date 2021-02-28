package main

import (
	"time"

	"github.com/Jleagle/rate-limit-go"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/queue"
	influx "github.com/influxdata/influxdb1-client"
	"go.uber.org/zap"
)

var (
	limits         = rate.New(time.Second*3, rate.WithBurst(3))
	discordSession *discordgo.Session
)

func main() {

	err := config.Init(helpers.GetIP())
	log.InitZap(log.LogNameChatbot)
	defer log.Flush()
	if err != nil {
		log.ErrS(err)
		return
	}

	if config.IsConsumer() {
		log.Err("Prod & local only")
		return
	}

	// Profiling
	// if config.IsLocal() {
	// 	go func() {
	// 		err := http.ListenAndServe(":6061", nil)
	// 		if err != nil {
	// 			log.ErrS(err)
	// 		}
	// 	}()
	// }

	err = mysql.GetConsumer("chatbot")
	if err != nil {
		log.ErrS(err)
		return
	}

	queue.Init(queue.ChatbotDefinitions)

	err = websocketServer()
	if err != nil {
		log.FatalS(err)
	}

	err = refreshCommands()
	if err != nil {
		log.Err("refreshing commands", zap.Error(err))
	}

	go updateGuildsCount()

	helpers.KeepAlive(
		mysql.Close,
		mongo.Close,
		func() {
			err = discordSession.Close()
			if err != nil {
				log.Err("disconnecting from discord", zap.Error(err))
			}
		},
		func() {
			influxHelper.GetWriter().Flush()
		},
	)
}

func updateGuildsCount() {

	if config.IsProd() {
		for {
			func() {
				var after = ""
				var count = 0
				for {

					guilds, err := discordSession.UserGuilds(100, "", after)
					if err != nil {
						log.ErrS(err)
						return
					}

					for _, guild := range guilds {
						count++
						after = guild.ID
					}

					if len(guilds) < 100 {
						break
					}
				}

				// Save to Influx
				point := influx.Point{
					Measurement: influxHelper.InfluxMeasurementChatBot.String(),
					Fields: map[string]interface{}{
						"guilds": count,
					},
					Precision: "h",
				}

				_, err := influxHelper.InfluxWrite(influxHelper.InfluxRetentionPolicyAllTime, point)
				if err != nil {
					log.ErrS(err)
				}
			}()

			time.Sleep(time.Hour)
		}
	}
}
