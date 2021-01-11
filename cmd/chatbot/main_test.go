package main

import (
	"testing"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/chatbot"
	"github.com/gamedb/gamedb/pkg/config"
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
		"app 440":          chatbot.CApp,
		"app tf2":          chatbot.CApp,
		"game 440":         chatbot.CApp,
		"game tf2":         chatbot.CApp,
		"new":              chatbot.CAppsNew,
		"players tf2":      chatbot.CAppPlayers,
		"online tf2":       chatbot.CAppPlayers,
		"popular":          chatbot.CAppsPopular,
		"random":           chatbot.CAppsRandom,
		"trending":         chatbot.CAppsTrending,
		"group tf2":        chatbot.CGroup,
		"clan tf2":         chatbot.CGroup,
		"trendinggroups":   chatbot.CGroupsTrending,
		"trending-groups":  chatbot.CGroupsTrending,
		"trending groups":  chatbot.CGroupsTrending,
		"help":             chatbot.CHelp,
		"players":          chatbot.CSteamOnline,
		"games Jleagle":    chatbot.CPlayerApps,
		"level Jleagle":    chatbot.CPlayerLevel,
		"player Jleagle":   chatbot.CPlayer,
		"playtime Jleagle": chatbot.CPlayerPlaytime,
		"recent Jleagle":   chatbot.CPlayerRecent,
		"update":           chatbot.CPlayerUpdate,
		"update Jleagle":   chatbot.CPlayerUpdate,
	}

	for _, start := range []string{".", "!"} {
		for message, commandID := range tests {
			for _, c := range chatbot.CommandRegister {
				if chatbot.RegexCache[c.Regex()].MatchString(start + message) {

					t.Log(start + message)

					if c.ID() != commandID {
						t.Error(c.ID(), "!=", commandID)
						continue
					}

					msg, err := c.Output("123", steamapi.ProductCCUS, c.LegacyInputs(start+message))
					if err != nil {
						t.Error(err)
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
