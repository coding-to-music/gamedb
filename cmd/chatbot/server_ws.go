package main

import (
	"strings"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/chatbot"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/discord"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mysql"
	"go.uber.org/zap"
)

func websocketServer() (*discordgo.Session, error) {

	// Start discord
	discordSession, err := getSession()
	if err != nil {
		return nil, err
	}

	// On joining a new guild
	discordSession.AddHandler(func(s *discordgo.Session, m *discordgo.GuildCreate) {

		err := memcache.Delete(memcache.ItemChatBotGuildsCount.Key)
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

				func() { // In a func for the defer

					// Disable PMs
					private, err := isChannelPrivateMessage(s, m)
					if err != nil {
						discordError(err)
						return
					}
					if private && m.Author.ID != discord.AdminID {
						return
					}

					// Save stats
					defer saveToDB(
						command,
						command.LegacyInputs(msg),
						false,
						m.GuildID,
						m.ChannelID,
						m.Author.ID,
						m.Author.Username,
						m.Author.Avatar,
					)

					// Typing notification
					err = s.ChannelTyping(m.ChannelID)
					discordError(err)

					// Get user settings
					code := steamapi.ProductCCUS
					if command.PerProdCode() {
						settings, err := mysql.GetChatBotSettings(m.Author.ID)
						if err != nil {
							log.ErrS(err)
						}
						code = settings.ProductCode
					}

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

	log.Info("Starting chatbot websocket connection")

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
