package main

import (
	"strings"
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/chatbot"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/discord"
	influxHelper "github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/ratelimit"
	"github.com/gamedb/gamedb/pkg/websockets"
	influx "github.com/influxdata/influxdb1-client"
	"go.uber.org/zap"
)

func websocketServer() (*discordgo.Session, error) {

	limits := ratelimit.New(time.Second, 5)

	// Start discord
	discordSession, err := discordgo.New("Bot " + config.C.DiscordChatBotToken)
	if err != nil {
		return nil, err
	}

	// On joining a new guild
	discordSession.AddHandlerOnce(func(s *discordgo.Session, m *discordgo.GuildCreate) {

		err := memcache.Delete(memcache.MemcacheChatBotGuildsCount.Key)
		if err != nil {
			log.ErrS(err)
			return
		}
	})

	// On new messages
	discordSession.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {

		// Don't reply to bots
		if m.Author.Bot {
			return
		}

		// Stop users getting two responses
		if config.IsLocal() && m.Author.ID != discord.AdminID {
			return
		}

		// Scan commands
		for _, command := range chatbot.CommandRegister {

			msg := strings.TrimSpace(m.Message.Content)

			if chatbot.RegexCache[command.Regex()].MatchString(msg) {

				// Disable PMs
				private, err := isPrivateChannel(s, m)
				if err != nil {
					discordError(err)
					return
				}
				if private && m.Author.ID != discord.AdminID {
					return
				}

				// Save stats
				//goland:noinspection GoDeferInLoop
				defer saveToDB(m, command, msg)

				// Typing notification
				err = discordSession.ChannelTyping(m.ChannelID)
				discordError(err)

				// React to request message
				// go func() {
				// 	err = discordSession.MessageReactionAdd(m.ChannelID, m.Message.ID, "👍")
				// 	discordError(err)
				// }()

				// Get user settings
				code := steamapi.ProductCCUS
				cacheItem := memcache.MemcacheChatBotRequest(msg, code)
				if command.PerProdCode() {
					settings, err := mysql.GetChatBotSettings(m.Author.ID)
					if err != nil {
						log.ErrS(err)
					}
					code = settings.ProductCode
					cacheItem = memcache.MemcacheChatBotRequest(msg, code)
				}

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
				if !limits.GetLimiter(m.Author.ID).Allow() {
					log.Warn("over chatbot rate limit", zap.String("author", m.Author.ID), zap.String("msg", msg))
					return
				}

				message, err := command.Output(m, code)
				if err != nil {
					log.WarnS(err, msg)
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
					log.ErrS(err, msg)
				}

				return
			}
		}
	})

	log.Info("Starting chatbot websocket connection")

	return discordSession, discordSession.Open()
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

func saveToDB(m *discordgo.MessageCreate, command chatbot.Command, message string) {

	if m.Author.ID == discord.AdminID {
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
		log.ErrS(err)
	}

	// Mongo
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