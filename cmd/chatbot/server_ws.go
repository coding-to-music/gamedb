package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/rate-limit-go"
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/chatbot"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/consumers"
	influxHelper "github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/websockets"
	influx "github.com/influxdata/influxdb1-client"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

var (
	rateLimit        = rate.New(time.Second*3, rate.WithBurst(3))
	errDirectMessage = "This command needs to be requested from a guild channel"
)

func websocketServer() (session *discordgo.Session, err error) {

	session, err = discordgo.New("Bot " + config.C.DiscordChatBotToken)
	if err != nil {
		return nil, err
	}

	session.AddHandler(func(s *discordgo.Session, e *discordgo.InteractionCreate) {

		// Check for pings
		if e.Type == discordgo.InteractionPing {

			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponsePong,
			}

			err := s.InteractionRespond(e.Interaction, response)
			if err != nil {
				log.ErrS(err)
			}
			return
		}

		// Get command
		command, ok := chatbot.CommandCache[e.Data.Name]
		if !ok {
			log.ErrS("Command ID not found in register")
			return
		}

		// Ignore PMs
		if !command.AllowDM() && e.User != nil {

			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionApplicationCommandResponseData{
					Content: errDirectMessage,
				},
			}

			err := s.InteractionRespond(e.Interaction, response)
			if err != nil {
				log.ErrS(err)
			}
			return
		}

		// Get user
		var user = e.User
		if user == nil {
			user = e.Member.User
		}

		// Save stats
		var success bool
		defer saveToDB(command, true, &success, argumentsString(e), e.GuildID, e.ChannelID, user)

		// Send an ACK response to update later
		err := s.InteractionRespond(e.Interaction, &discordgo.InteractionResponse{Type: discordgo.InteractionResponseDeferredChannelMessageWithSource})
		if err != nil {
			log.ErrS(err)
			return
		}

		// Check in cache first
		code := getProdCC(command, user.ID)
		cacheItem := memcache.ItemChatBotRequestSlash(command.ID(), arguments(e), code)

		if !command.DisableCache() && !config.IsLocal() {

			var edit = &discordgo.WebhookEdit{}
			err = memcache.Client().Get(cacheItem.Key, edit)
			if err == nil {

				err = s.InteractionResponseEdit(config.C.DiscordChatBotClientID, e.Interaction, edit)
				if err != nil {
					log.ErrS(err)
				}
				return
			}
		}

		// Rate limit
		if !rateLimit.GetLimiter(user.ID).Allow() {
			log.Warn("over chatbot rate limit", zap.String("author", user.ID), zap.String("msg", argumentsString(e)))
			return
		}

		// Make output
		out, err := command.Output(user.ID, code, arguments(e))
		if err != nil {
			log.ErrS(err)
			return
		}

		// Convert to edit
		edit := &discordgo.WebhookEdit{
			Content: out.Content,
		}

		if out.Embed != nil {
			edit.Embeds = []*discordgo.MessageEmbed{out.Embed}
		}

		// Respond
		err = s.InteractionResponseEdit(config.C.DiscordChatBotClientID, e.Interaction, edit)
		if err != nil {
			log.ErrS(err)
		}

		// Save to cache
		err = memcache.Client().Set(cacheItem.Key, edit, cacheItem.Expiration)
		if err != nil {
			log.Err("Saving to memcache", zap.Error(err), zap.String("msg", argumentsString(e)))
		}

		success = true
	})

	session.AddHandler(func(s *discordgo.Session, e *discordgo.MessageCreate) {

		// Don't reply to bots
		if e.Author.Bot {
			return
		}

		// Scan commands
		for _, command := range chatbot.CommandRegister {

			msg := strings.TrimSpace(e.Content)

			if chatbot.RegexCache[command.Regex()].MatchString(msg) {

				func() { // In a func for the defer

					// Ignore PMs
					private := func() bool {
						channel, err := s.State.Channel(e.ChannelID)
						if err != nil {
							channel, err = s.Channel(e.ChannelID)
							if err != nil {
								discordError(err)
								return false
							}
						}
						return channel.Type == discordgo.ChannelTypeDM
					}()

					if !command.AllowDM() && private {

						message := discordgo.MessageSend{
							Content: errDirectMessage,
						}

						err = sendMessage(s, e, msg, &message)
						discordError(err)
						return
					}

					// Save stats
					var success bool
					defer saveToDB(command, false, &success, msg, e.GuildID, e.ChannelID, e.Author)

					// Typing notification
					err := s.ChannelTyping(e.ChannelID)
					discordError(err)

					// Get user settings
					code := getProdCC(command, e.Author.ID)

					cacheItem := memcache.ItemChatBotRequest(msg, code)

					// Check in cache first
					if !command.DisableCache() && !config.IsLocal() {
						var message discordgo.MessageSend
						err = memcache.Client().Get(cacheItem.Key, &message)
						if err == nil {

							err = sendMessage(s, e, msg, &message)
							if err != nil {
								discordError(err)
								return
							}

							success = true
							return
						}
					}

					// Rate limit
					if !rateLimit.GetLimiter(e.Author.ID).Allow() {
						log.Warn("over chatbot rate limit", zap.String("author", e.Author.ID), zap.String("msg", msg))
						success = true
						return
					}

					// Make output
					message, err := command.Output(e.Author.ID, code, command.LegacyInputs(msg))
					if err != nil {
						log.ErrS(err, msg)
						return
					}

					// Save to cache
					defer func() {
						err = memcache.Client().Set(cacheItem.Key, message, cacheItem.Expiration)
						if err != nil {
							log.ErrS(err, msg)
						}
					}()

					// Reply
					err = sendMessage(s, e, msg, &message)
					if err != nil {
						discordError(err)
						return
					}

					success = true
				}()

				break
			}
		}
	})

	// When the bot joins a guild
	session.AddHandler(func(_ *discordgo.Session, e *discordgo.GuildCreate) {

		if e.MemberCount == 0 {
			return
		}

		mongoGuild := mongo.DiscordGuild{
			ID:      e.ID,
			Name:    e.Name,
			Icon:    e.Icon,
			Members: e.MemberCount,
		}

		_, err := mongo.ReplaceOne(mongo.CollectionDiscordGuilds, bson.D{{"_id", e.ID}}, mongoGuild)
		if err != nil {
			log.Err("Updating guild row", zap.Error(err))
		}
	})

	log.Info("Starting chatbot websocket connection")
	err = session.Open()
	if err != nil {
		return nil, err
	}

	return session, nil
}

func getProdCC(command chatbot.Command, authorID string) steamapi.ProductCC {

	code := steamapi.ProductCCUS
	if command.PerProdCode() {
		settings, err := mysql.GetChatBotSettings(authorID)
		if err != nil {
			log.ErrS(err)
		}
		code = settings.ProductCode
	}
	return code
}

func saveToDB(command chatbot.Command, isSlash bool, wasSuccess *bool, message, guildID, channelID string, user *discordgo.User) {

	if config.IsLocal() {
		return
	}
	if user.ID == config.DiscordAdminID {
		return
	}
	if !isSlash && (command.ID() == chatbot.CHelp || command.ID() == chatbot.CInvite) {
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
			"success":    strconv.FormatBool(*wasSuccess),
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

	var row = mongo.DiscordMessage{
		GuildID:      guildID,
		ChannelID:    channelID,
		AuthorID:     user.ID,
		AuthorName:   user.Username,
		AuthorAvatar: user.Avatar,
		CommandID:    command.ID(),
		Message:      message,
		Slash:        isSlash,
	}

	_, err = mongo.InsertOne(mongo.CollectionChatBotCommands, row)
	if err != nil {
		log.ErrS(err)
	}

	// Websocket
	guilds, err := mongo.GetGuildsByIDs([]string{row.GuildID})
	if err != nil {
		log.ErrS(err)
	}

	wsPayload := consumers.ChatBotPayload{}
	wsPayload.RowData = row.GetTableRowJSON(guilds)

	err = consumers.ProduceWebsocket(wsPayload, websockets.PageChatBot)
	if err != nil {
		log.ErrS(err)
	}
}

func sendMessage(s *discordgo.Session, event *discordgo.MessageCreate, messageRaw string, message *discordgo.MessageSend) (err error) {

	_, err = s.ChannelMessageSendComplex(event.ChannelID, message)
	if err != nil {

		// If reply failed, try to DM the OP
		if val, ok := err.(*discordgo.RESTError); ok && val.Message.Code == 50013 { // Missing Permissions

			channel, err := s.UserChannelCreate(event.Author.ID)
			if err != nil {
				log.Err("Getting user channel", zap.Error(err), zap.String("msg", messageRaw))
				return err
			}

			_, err = s.ChannelMessageSend(channel.ID, "I do not have permission to post in that channel :(")
			if err != nil {

				if _, ok := err.(*discordgo.RESTError); ok && val.Response.StatusCode == 403 {
					return nil
				}

				log.Err("Sending channel message", zap.Error(err), zap.String("msg", messageRaw))
				return err
			}
		}
	}

	return err
}

func arguments(event *discordgo.InteractionCreate) (a map[string]string) {

	a = map[string]string{}
	for _, v := range event.Data.Options {
		a[v.Name] = fmt.Sprint(v.Value)
	}
	return a
}

func argumentsString(event *discordgo.InteractionCreate) string {

	var s = []string{event.Data.Name}
	for _, v := range event.Data.Options {
		s = append(s, fmt.Sprint(v.Value))
	}
	return strings.Join(s, " ")
}

func discordError(err error) {

	if err != nil {
		if val, ok := err.(*discordgo.RESTError); ok {

			var allowed = map[int]string{
				50001: "Missing Access",
				50013: "Missing Permissions",
			}

			if _, ok2 := allowed[val.Message.Code]; ok2 {
				zap.S().Info(err) // No helper to fix stack offset
				return
			}
		}

		zap.S().Error(err) // No helper to fix stack offset
	}
}
