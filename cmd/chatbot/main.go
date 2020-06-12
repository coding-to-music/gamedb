package main

import (
	"net/http"
	_ "net/http/pprof"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/didip/tollbooth/v5"
	"github.com/didip/tollbooth/v5/limiter"
	"github.com/gamedb/gamedb/pkg/chatbot"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/helpers/influx"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/websockets"
	influx "github.com/influxdata/influxdb1-client"
)

const debugAuthorID = "145456943912189952"

var (
	version string

	ignoreGuildIDs = []string{
		// "110373943822540800", // Discord Bots
	}
)

func main() {

	config.Init(version, helpers.GetIP())
	log.Initialise(log.LogNameChatbot)

	log.Info("Starting chatbot")

	// Load PPROF
	if config.IsLocal() {
		log.Info("Starting chatbot profiling")
		go func() {
			err := http.ListenAndServe("localhost:6062", nil)
			log.Critical(err)
		}()
	}

	// Get API key
	err := sql.GetAPIKey("chatbot")
	if err != nil {
		log.Critical(err)
		return
	}

	if !config.IsProd() && !config.IsLocal() {
		log.Err("Prod & local only")
		return
	}

	// Load consumers
	queue.Init(queue.ChatbotDefinitions)

	// Cache regexes
	for _, v := range chatbot.CommandRegister {
		chatbot.RegexCache[v.Regex()] = regexp.MustCompile(v.Regex())
	}

	// Set limiter
	ops := limiter.ExpirableOptions{DefaultExpirationTTL: time.Second}
	lmt := limiter.New(&ops).SetMax(1).SetBurst(5)

	// Start discord
	discordSession, err := discordgo.New("Bot " + config.Config.DiscordChatBotToken.Get())
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
				if m.Author.ID != debugAuthorID && !helpers.SliceHasString(m.GuildID, ignoreGuildIDs) {
					//noinspection GoDeferInLoop
					defer func() {
						go saveToInflux(m, command)
						go saveToMongo(m, msg)
					}()
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
				if !command.DisableCache() {
					var message discordgo.MessageSend
					err = memcache.GetInterface(cacheItem.Key, &message)
					if err == nil {
						_, err = s.ChannelMessageSendComplex(m.ChannelID, &message)
						discordError(err)
						return
					}
				}

				// Rate limit
				httpErr := tollbooth.LimitByKeys(lmt, []string{m.Author.ID})
				if httpErr != nil {
					log.Warning(m.Author.ID + " over chatbot rate limit")
					return
				}

				message, err := command.Output(m)
				if err != nil {
					log.Warning(err, msg)
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
					log.Err(err, msg)
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
			"command":    command.Regex(),
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

	err := mongo.CreateChatBotCommand(*m, message)
	if err != nil {
		log.Err(err)
		return
	}

	wsPayload := queue.ChatBotPayload{}
	wsPayload.AuthorID = m.Author.ID
	wsPayload.AuthorName = m.Author.Username
	wsPayload.AuthorAvatar = m.Author.Avatar
	wsPayload.Message = message

	err = queue.ProduceWebsocket(wsPayload, websockets.PageChatBot)
	if err != nil {
		log.Err(err)
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
				log.Info(err)
				return
			}
		}

		log.Err(err)
	}
}
