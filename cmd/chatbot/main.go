package main

import (
	"net/http"
	_ "net/http/pprof"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/didip/tollbooth/v6"
	"github.com/didip/tollbooth/v6/limiter"
	"github.com/gamedb/gamedb/pkg/chatbot"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/websockets"
	influx "github.com/influxdata/influxdb1-client"
	"go.uber.org/zap"
)

const debugAuthorID = "145456943912189952"

var version string
var commits string

//noinspection GoDeferInLoop
func main() {

	config.Init(version, commits, helpers.GetIP())
	log.InitZap(log.LogNameChatbot)

	zap.L().Info("Starting chatbot")

	// Profiling
	if !config.IsConsumer() {
		zap.L().Info("Starting chatbot profiling")
		go func() {
			err := http.ListenAndServe(":6061", nil)
			if err != nil {
				zap.S().Fatal(err)
			}
		}()
	}

	// Get API key
	err := mysql.GetConsumer("chatbot")
	if err != nil {
		zap.S().Fatal(err)
		return
	}

	if !config.IsProd() && !config.IsLocal() {
		zap.L().Error("Prod & local only")
		return
	}

	// Load consumers
	queue.Init(queue.ChatbotDefinitions)

	// Set limiter
	ops := limiter.ExpirableOptions{DefaultExpirationTTL: time.Second}
	lmt := limiter.New(&ops).SetMax(1).SetBurst(5)

	// Start discord
	discordSession, err := discordgo.New("Bot " + config.C.DiscordChatBotToken)
	if err != nil {
		panic("Can't create Discord session")
	}

	discordSession.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {

		// Don't reply to bots
		if m.Author.Bot {
			return
		}

		// Stop users getting two responses
		if config.IsLocal() && m.Author.ID != debugAuthorID {
			return
		}

		// Scan commands
		for _, command := range chatbot.CommandRegister {

			msg := strings.TrimSpace(m.Message.Content)

			if chatbot.RegexCache[command.Regex()].MatchString(msg) {

				cacheItem := memcache.MemcacheChatBotRequest(msg)

				// Disable PMs
				private, err := isPrivateChannel(s, m)
				if err != nil {
					discordError(err)
					return
				}
				if private && m.Author.ID != debugAuthorID {
					return
				}

				// Save stats
				if m.Author.ID != debugAuthorID {
					defer saveToInflux(m, command)
					defer saveToMongo(m, command, msg)
				}

				// Typing notification
				err = discordSession.ChannelTyping(m.ChannelID)
				discordError(err)

				// React to request message
				// go func() {
				// 	err = discordSession.MessageReactionAdd(m.ChannelID, m.Message.ID, "👍")
				// 	discordError(err)
				// }()

				// Check in cache first
				if !command.DisableCache() && !config.IsLocal() {
					var message discordgo.MessageSend
					err = memcache.GetInterface(cacheItem.Key, &message)
					if err == nil {
						_, err = s.ChannelMessageSendComplex(m.ChannelID, &message)
						discordError(err)
						return
					}
				}

				// Rate limit
				err = tollbooth.LimitByKeys(lmt, []string{m.Author.ID})
				if err != nil {
					zap.L().Warn("over chatbot rate limit", zap.String("author", m.Author.ID), zap.String("msg", msg))
					return
				}

				message, err := command.Output(m)
				if err != nil {
					zap.S().Warn(err, msg)
					return
				}

				_, err = s.ChannelMessageSendComplex(m.ChannelID, &message)
				if err != nil {
					discordError(err)
					return
				}

				// Save to cache
				err = memcache.SetInterface(cacheItem.Key, message, cacheItem.Expiration)
				if err != nil {
					zap.S().Error(err, msg)
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

	point := influx.Point{
		Measurement: string(influxHelper.InfluxMeasurementChatBot),
		Tags: map[string]string{
			"guild_id":   m.GuildID,
			"channel_id": m.ChannelID,
			"author_id":  m.Author.ID,
			"command":    command.Regex(),
			"command_id": command.ID(),
		},
		Fields: map[string]interface{}{
			"request": 1,
		},
		Time:      time.Now(),
		Precision: "ms",
	}

	_, err := influxHelper.InfluxWrite(influxHelper.InfluxRetentionPolicyAllTime, point)
	if err != nil {
		zap.S().Error(err)
	}
}

func saveToMongo(m *discordgo.MessageCreate, command chatbot.Command, message string) {

	if config.IsLocal() {
		return
	}

	t, _ := m.Timestamp.Parse()

	var row = mongo.ChatBotCommand{
		GuildID:      m.GuildID,
		ChannelID:    m.ChannelID,
		AuthorID:     m.Author.ID,
		AuthorName:   m.Author.Username,
		AuthorAvatar: m.Author.Avatar,
		CommandID:    command.ID(),
		Message:      message,
		Time:         t,
	}

	_, err := mongo.InsertOne(mongo.CollectionChatBotCommands, row)
	if err != nil {
		zap.S().Error(err)
		return
	}

	wsPayload := queue.ChatBotPayload{}
	wsPayload.AuthorID = m.Author.ID
	wsPayload.AuthorName = m.Author.Username
	wsPayload.AuthorAvatar = m.Author.Avatar
	wsPayload.Message = message

	err = queue.ProduceWebsocket(wsPayload, websockets.PageChatBot)
	if err != nil {
		zap.S().Error(err)
		return
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
				zap.S().Info(err)
				return
			}
		}

		zap.S().Error(err)
	}
}
