package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/rate-limit-go"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/chatbot"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/websockets"
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

func saveToDB(command chatbot.Command, isSlash, wasSuccess bool, message, guildID, channelID string, user *discordgo.User) {

	if user.ID == config.DiscordAdminID {
		return
	}

	if config.IsLocal() {
		return
	}

	if command.ID() == chatbot.CHelp {
		return
	}

	// Influx
	point := influx.Point{
		Measurement: string(influxHelper.InfluxMeasurementChatBot),
		Tags: map[string]string{
			"guild_id":   guildID,
			"channel_id": channelID,
			"author_id":  user.ID,
			"command_id": command.ID(),
			"slash":      strconv.FormatBool(isSlash),
			"success":    strconv.FormatBool(wasSuccess),
		},
		Fields: map[string]interface{}{
			"request": 1,
		},
		Time:      time.Now(),
		Precision: "ms",
	}

	_, err := influxHelper.InfluxWrite(influxHelper.InfluxRetentionPolicyAllTime, point)
	if err != nil {
		log.ErrS(err)
	}

	var row = mongo.ChatBotCommand{
		GuildID:      guildID,
		ChannelID:    channelID,
		AuthorID:     user.ID,
		AuthorName:   user.Username,
		AuthorAvatar: user.Avatar,
		CommandID:    command.ID(),
		Message:      message,
		Slash:        isSlash,
		Time:         time.Now(), // Can get from ws message?
	}

	_, err = mongo.InsertOne(mongo.CollectionChatBotCommands, row)
	if err != nil {
		log.ErrS(err)
	}

	// Websocket
	guilds, err := mongo.GetGuilds([]string{row.GuildID})
	if err != nil {
		log.ErrS(err)
	}

	wsPayload := queue.ChatBotPayload{}
	wsPayload.RowData = row.GetTableRowJSON(guilds)

	err = queue.ProduceWebsocket(wsPayload, websockets.PageChatBot)
	if err != nil {
		log.ErrS(err)
	}
}

func discordError(err error) {

	var allowed = map[int]string{
		50001: "Missing Access",
		50013: "Missing Permissions",
	}

	if err != nil {
		if val, ok := err.(*discordgo.RESTError); ok {
			if _, ok2 := allowed[val.Message.Code]; ok2 {
				zap.S().Info(err) // No helper to fix stack offset
				return
			}
		}

		zap.S().Error(err) // No helper to fix stack offset
		return
	}
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

func arguments(event *discordgo.Interaction) (a map[string]string) {

	a = map[string]string{}
	for _, v := range event.Data.Options {
		a[v.Name] = fmt.Sprint(v.Value)
	}
	return a
}

func argumentsString(event *discordgo.Interaction) string {

	var s = []string{event.Data.Name}
	for _, v := range event.Data.Options {
		s = append(s, fmt.Sprint(v.Value))
	}
	return strings.Join(s, " ")
}
