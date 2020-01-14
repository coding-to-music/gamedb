package main

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/didip/tollbooth/limiter"
	"github.com/didip/tollbooth/v5"
	"github.com/gamedb/gamedb/pkg/chatbot"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/helpers/influx"
	pubsubHelpers "github.com/gamedb/gamedb/pkg/helpers/pubsub"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/websockets"
	influx "github.com/influxdata/influxdb1-client"
)

const debugAuthorID = "145456943912189952"

var (
	version string
)

func main() {

	config.SetVersion(version)
	log.Initialise([]log.LogName{log.LogNameChatbot})

	// Get API key
	err := sql.GetAPIKey("chatbot")
	if err != nil {
		log.Critical(err)
		return
	}

	log.Info("Starting chatbot")

	if !config.IsProd() && !config.IsLocal() {
		log.Err("Prod & local only")
	}

	ops := limiter.ExpirableOptions{DefaultExpirationTTL: time.Second}
	lmt := limiter.New(&ops).SetMax(1).SetBurst(5)

	//
	discordSession, err := discordgo.New("Bot " + config.Config.DiscordChatBotToken.Get())
	if err != nil {
		panic("Can't create Discord session")
	}

	discordSession.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {

		// Don't reply to bots
		if m.Author.Bot {
			return
		}

		// Rate limit
		httpErr := tollbooth.LimitByKeys(lmt, []string{m.Author.ID})
		if httpErr != nil {
			log.Warning(m.Author.ID + " over rate limit")
			return
		}

		// Scan commands
		for _, command := range chatbot.CommandRegister {

			msg := m.Message.Content

			if command.Regex().MatchString(msg) {

				go saveToInflux(m, command)
				go saveToMongo(m, msg)

				go func() {
					err := discordSession.ChannelTyping(m.ChannelID)
					log.Err(err)
				}()

				chanID := m.ChannelID

				// Allow private messaging for admins
				if m.Author.ID == debugAuthorID {

					private, err := isPrivateChannel(s, m)
					if err != nil {
						log.Warning(err, msg)
						return
					}

					if private {

						st, err := s.UserChannelCreate(m.Author.ID)
						if err != nil {
							log.Warning(err, msg)
							return
						}

						chanID = st.ID
					}
				}

				message, err := command.Output(msg)
				if err != nil {
					log.Warning(err, msg)
					return
				}

				_, err = s.ChannelMessageSendComplex(chanID, &message)
				if err != nil {
					log.Warning(err, msg)
					return
				}

				return
			}
		}
	})

	err = discordSession.Open()
	if err != nil {
		panic("Can't connect to Discord session")
	}

	helpers.KeepAlive()
}

func isPrivateChannel(s *discordgo.Session, m *discordgo.MessageCreate) (bool, error) {
	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		if channel, err = s.Channel(m.ChannelID); err != nil {
			return false, err
		}
	}

	return channel.Type == discordgo.ChannelTypeDM, nil
}

func saveToInflux(m *discordgo.MessageCreate, command chatbot.Command) {

	if config.IsLocal() {
		return
	}

	_, err := influxHelper.InfluxWrite(influxHelper.InfluxRetentionPolicyAllTime, influx.Point{
		Measurement: string(influxHelper.InfluxMeasurementChatBot),
		Tags: map[string]string{
			"guild_id":   m.GuildID,
			"channel_id": m.ChannelID,
			"author_id":  m.Author.ID,
			"command":    command.Regex().String(),
		},
		Fields: map[string]interface{}{
			"request": 1,
		},
		Time:      time.Now(),
		Precision: "u",
	})
	log.Err(err)
}

func saveToMongo(m *discordgo.MessageCreate, message string) {

	if config.IsLocal() {
		return
	}

	var command = mongo.ChatBotCommand{
		Guild:   m.GuildID,
		Channel: m.ChannelID,
		Author:  m.Author.ID,
		Message: message,
	}

	_, err := mongo.InsertOne(mongo.CollectionChatBotCommands, command)
	if err != nil {
		log.Err(err)
		return
	}

	wsPayload := websockets.PubSubStringPayload{}
	wsPayload.String = message
	wsPayload.Pages = []websockets.WebsocketPage{websockets.PageChatBot}

	_, err = pubsubHelpers.Publish(pubsubHelpers.PubSubTopicWebsockets, wsPayload)
	if err != nil {
		log.Err(err)
	}
}
