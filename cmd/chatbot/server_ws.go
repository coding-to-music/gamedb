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

func websocketServer() (*discordgo.Session, error) {

	// Start discord
	discordSession, err := discordgo.New("Bot " + config.C.DiscordChatBotToken)
	if err != nil {
		return nil, err
	}

	discordSession.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {

		interaction := i.Interaction

		// Stop users getting two responses
		if config.IsLocal() && interaction.Member.User.ID != config.DiscordAdminID {
			ackWithSource(discordSession, interaction)
			return
		}

		// Check for pings
		if interaction.Type == discordgo.InteractionPing {

			response := &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponsePong,
			}

			err = discordSession.InteractionRespond(interaction, response)
			if err != nil {
				log.ErrS(err)
			}
			return
		}

		// Get command
		command, ok := chatbot.CommandCache[interaction.Data.Name]
		if !ok {
			ackWithSource(s, interaction)
			log.ErrS("Command ID not found in register")
			return
		}

		// Save stats
		defer saveToDB(command, true, argumentsString(interaction), interaction.GuildID, interaction.ChannelID, interaction.Member.User)

		// Typing notification
		err = discordSession.ChannelTyping(interaction.ChannelID)
		discordError(err)

		ackWithSource(discordSession, interaction)

		//
		code := getAuthorCode(command, interaction.Member.User.ID)

		cacheItem := memcache.ItemChatBotRequestSlash(command.ID(), arguments(interaction), code)

		// Check in cache first
		if !command.DisableCache() && !config.IsLocal() {

			var response = &discordgo.InteractionResponse{}
			err = memcache.GetInterface(cacheItem.Key, &response)
			if err == nil {

				err = discordSession.InteractionRespond(interaction, response)
				if err != nil {
					log.ErrS(err)
				}
				return
			}
		}

		// Rate limit
		if !limits.GetLimiter(interaction.Member.User.ID).Allow() {
			log.Warn("over chatbot rate limit", zap.String("author", interaction.Member.User.ID), zap.String("msg", argumentsString(interaction)))
			ackWithSource(discordSession, interaction)
			return
		}

		out, err := command.Output(interaction.Member.User.ID, code, arguments(interaction))
		if err != nil {
			log.ErrS(err)
			ackWithSource(discordSession, interaction)
			return
		}

		response := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionApplicationCommandResponseData{
				Content: out.Content,
			},
		}

		if out.Embed != nil {
			response.Data.Embeds = []*discordgo.MessageEmbed{out.Embed}
		}

		// Save to cache
		defer func() {
			err = memcache.SetInterface(cacheItem.Key, response, cacheItem.Expiration)
			if err != nil {
				log.Err("Saving to memcache", zap.Error(err), zap.String("msg", argumentsString(interaction)))
			}
		}()

		update := &discordgo.WebhookEdit{
			Content: response.Data.Content,
			Embeds:  response.Data.Embeds,
		}

		err = discordSession.InteractionResponseEdit("", interaction, update)
		if err != nil {
			log.ErrS(err)
		}
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

					// Disable PMs
					private, err := isChannelPrivateMessage(s, m)
					if err != nil {
						discordError(err)
						return
					}
					if private && m.Author.ID != config.DiscordAdminID {
						return
					}

					// Save stats
					defer saveToDB(command, false, msg, m.GuildID, m.ChannelID, m.Author)

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

					message, err := command.Output(m.Author.ID, code, command.LegacyInputs(msg))
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
				}()

				break
			}
		}
	})

	return discordSession, discordSession.Open()
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

func ackWithSource(s *discordgo.Session, i *discordgo.Interaction) {

	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseACKWithSource,
	}

	err := s.InteractionRespond(i, response)
	if err != nil {
		log.ErrS(err)
	}
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
