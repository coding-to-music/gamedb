package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/website/pkg/chatbot"
	"github.com/gamedb/website/pkg/config"
	"github.com/gamedb/website/pkg/helpers"
	"github.com/gamedb/website/pkg/log"
)

const debugAuthorID = "145456943912189952"

func main() {

	log.Info("Starting chatbot")

	if !config.IsProd() && !config.IsLocal() {
		log.Err("Prod & local only")
	}

	discord, err := discordgo.New("Bot " + config.Config.DiscordBotToken.Get())
	if err != nil {
		fmt.Println(err)
		return
	}

	discord.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {

		// Don't reply to bots
		if m.Author.Bot {
			return
		}

		for _, command := range chatbot.CommandRegister {

			msg := m.Message.Content

			if command.Regex().MatchString(msg) {

				chanID := m.ChannelID

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

	err = discord.Open()
	if err != nil {
		fmt.Println(err)
		return
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
