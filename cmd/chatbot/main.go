package main

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/chatbot"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/discord"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/ratelimit"
	"github.com/gamedb/gamedb/pkg/websockets"
	influx "github.com/influxdata/influxdb1-client"
	"go.uber.org/zap"
)

var limits = ratelimit.New(time.Second, 3)

func main() {

	err := config.Init(helpers.GetIP())
	log.InitZap(log.LogNameChatbot)
	defer log.Flush()
	if err != nil {
		log.ErrS(err)
		return
	}

	if !config.IsConsumer() {
		go func() {
			err := http.ListenAndServe(":6061", nil)
			if err != nil {
				log.ErrS(err)
			}
		}()
	}

	err = mysql.GetConsumer("chatbot")
	if err != nil {
		log.ErrS(err)
		return
	}

	if config.IsConsumer() {
		log.Err("Prod & local only")
		return
	}

	queue.Init(queue.ChatbotDefinitions)

	discordSession, err := websocketServer()
	if err != nil {
		log.FatalS(err)
	}

	err = slashCommandServer()
	if err != nil {
		log.FatalS(err)
	}

	if config.IsProd() {

		err = refreshCommands()
		if err != nil {
			log.Err("refreshing commands", zap.Error(err))
		}
	}

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

var discordSession *discordgo.Session
var discordSessionLock sync.Mutex

func getSession() (*discordgo.Session, error) {

	discordSessionLock.Lock()
	defer discordSessionLock.Unlock()

	var err error

	if discordSession == nil {
		discordSession, err = discordgo.New("Bot " + config.C.DiscordChatBotToken)
		if err != nil {
			return nil, err
		}
	}

	return discordSession, err

}

func saveToDB(command chatbot.Command, slash bool, message, guildID, channelID, authorID, authorName, authorAvatar string) {

	if authorID == discord.AdminID {
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
			"author_id":  authorID,
			"command_id": command.ID(),
			"slash":      strconv.FormatBool(slash),
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
		AuthorID:     authorID,
		AuthorName:   authorName,
		AuthorAvatar: authorAvatar,
		CommandID:    command.ID(),
		Message:      message,
		Slash:        slash,
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
