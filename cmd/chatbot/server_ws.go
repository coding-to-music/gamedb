package main

import (
	"strings"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/chatbot"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mysql"
	"go.uber.org/zap"
)

func websocketServer() (err error) {

	// Start discord
	discordSession, err = discordgo.New("Bot " + config.C.DiscordChatBotToken)
	if err != nil {
		return err
	}

	discordSession.AddHandler(func(s *discordgo.Session, interaction *discordgo.InteractionCreate) {

		i := interaction.Interaction

		// Ignore PMs
		// member is sent when the command is invoked in a guild, and user is sent when invoked in a DM
		// todo, make PR to add user with isDM() func
		if i.Member == nil {
			return
		}

		// Stop users getting two responses
		if config.IsLocal() && i.Member.User.ID != config.DiscordAdminID {
			return
		}

		// Check for pings
		if i.Type == discordgo.InteractionPing {

			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponsePong,
			}

			err = s.InteractionRespond(i, response)
			if err != nil {
				log.ErrS(err)
			}
			return
		}

		// Get command
		command, ok := chatbot.CommandCache[i.Data.Name]
		if !ok {
			log.ErrS("Command ID not found in register")
			return
		}

		// Save stats
		var success bool
		defer saveToDB(command, true, success, argumentsString(i), i.GuildID, i.ChannelID, i.Member.User)

		// Typing notification
		// todo Remove this when slash commands have `thinking`
		err = s.ChannelTyping(i.ChannelID)
		discordError(err)

		//
		code := getAuthorCode(command, i.Member.User.ID)

		cacheItem := memcache.ItemChatBotRequestSlash(command.ID(), arguments(i), code)

		// Check in cache first
		if !command.DisableCache() && !config.IsLocal() {

			var response = &discordgo.InteractionResponse{}
			err = memcache.GetInterface(cacheItem.Key, &response)
			if err == nil {

				err = s.InteractionRespond(i, response)
				if err != nil {
					log.ErrS(err)
				}
				return
			}
		}

		// Rate limit
		if !limits.GetLimiter(i.Member.User.ID).Allow() {
			log.Warn("over chatbot rate limit", zap.String("author", i.Member.User.ID), zap.String("msg", argumentsString(i)))
			return
		}

		// Make output
		out, err := command.Output(i.Member.User.ID, code, arguments(i))
		if err != nil {
			log.ErrS(err)
			return
		}

		// Convert to slash format
		response := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionApplicationCommandResponseData{
				Content: out.Content,
			},
		}

		if out.Embed != nil {
			response.Data.Embeds = []*discordgo.MessageEmbed{out.Embed}
		}

		// Respond
		err = s.InteractionRespond(i, response)
		if err != nil {
			log.ErrS(err)
		}

		// Save to cache
		err = memcache.SetInterface(cacheItem.Key, response, cacheItem.Expiration)
		if err != nil {
			log.Err("Saving to memcache", zap.Error(err), zap.String("msg", argumentsString(i)))
		}

		success = true
	})

	// On new messages
	discordSession.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {

		// Don't reply to bots
		if m.Author.Bot {
			return
		}

		// Stop users getting two responses
		if config.IsLocal() && m.Author.ID != config.DiscordAdminID {
			return
		}

		// Scan commands
		for _, command := range chatbot.CommandRegister {

			msg := strings.TrimSpace(m.Message.Content)

			if chatbot.RegexCache[command.Regex()].MatchString(msg) {

				func() { // In a func for the defer

					// Ignore PMs
					private, err := isChannelPrivateMessage(s, m)
					if err != nil {
						discordError(err)
						return
					}
					if private && m.Author.ID != config.DiscordAdminID {
						return
					}

					// Save stats
					var success bool
					defer saveToDB(command, false, success, msg, m.GuildID, m.ChannelID, m.Author)

					// Typing notification
					err = s.ChannelTyping(m.ChannelID)
					discordError(err)

					// Get user settings
					code := getAuthorCode(command, m.Author.ID)

					cacheItem := memcache.ItemChatBotRequest(msg, code)

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

					// Make output
					message, err := command.Output(m.Author.ID, code, command.LegacyInputs(msg))
					if err != nil {
						log.WarnS(err, msg)
						return
					}

					// Reply
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

					success = true
				}()

				break
			}
		}
	})

	log.Info("Starting chatbot websocket connection")
	return discordSession.Open()
}

func isChannelPrivateMessage(s *discordgo.Session, m *discordgo.MessageCreate) (bool, error) {
	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		if channel, err = s.Channel(m.ChannelID); err != nil {
			return false, err
		}
	}

	return channel.Type == discordgo.ChannelTypeDM, nil
}

func getAuthorCode(command chatbot.Command, authorID string) steamapi.ProductCC {

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
