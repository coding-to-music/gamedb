package main

import (
	"testing"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/chatbot"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
)

func Test(t *testing.T) {

	err := config.Init("")
	log.InitZap(log.LogNameTest)
	defer log.Flush()
	if err != nil {
		log.ErrS(err)
		return
	}

	tests := map[string]string{
		"app 440":           chatbot.CApp,
		"app tf2":           chatbot.CApp,
		"game {game}":       chatbot.CApp,
		"new":               chatbot.CAppsNew,
		"players {game}":    chatbot.CAppPlayers,
		"online {game}":     chatbot.CAppPlayers,
		"popular":           chatbot.CAppsPopular,
		"random":            chatbot.CAppsRandom,
		"trending":          chatbot.CAppsTrending,
		"group {game}":      chatbot.CGroup,
		"clan {game}":       chatbot.CGroup,
		"trendinggroups":    chatbot.CGroupsTrending,
		"trending-groups":   chatbot.CGroupsTrending,
		"trending groups":   chatbot.CGroupsTrending,
		"help":              chatbot.CHelp,
		"players":           chatbot.CSteamOnline,
		"games {player}":    chatbot.CPlayerApps,
		"level {player}":    chatbot.CPlayerLevel,
		"player {player}":   chatbot.CPlayer,
		"playtime {player}": chatbot.CPlayerPlaytime,
		"recent {player}":   chatbot.CPlayerRecent,
		"update":            chatbot.CPlayerUpdate,
		"update {player}":   chatbot.CPlayerUpdate,
	}

	for _, start := range []string{".", "!"} {
		for message, commandID := range tests {
			for _, c := range chatbot.CommandRegister {
				if chatbot.RegexCache[c.Regex()].MatchString(start + message) {

					if c.ID() != commandID {
						t.Error(start+message, "!=", commandID)
					}

					if start == "." {

						t.Log(start + message)

						messagex := &discordgo.MessageCreate{
							Message: &discordgo.Message{
								Content: start + message,
								Author: &discordgo.User{
									ID: "123",
								},
							},
						}

						msg, err := c.Output(messagex, steamapi.ProductCCUS)
						if err != nil {
							t.Error(start+message, "!=", commandID)
							continue
						}
						if msg.Content == "" && msg.Embed == nil {
							t.Error("no return")
							continue
						}
					}
				}
			}
		}
	}
}
