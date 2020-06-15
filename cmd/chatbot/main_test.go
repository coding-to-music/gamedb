package main

import (
	"testing"

	"github.com/gamedb/gamedb/pkg/chatbot"
)

func Test(t *testing.T) {

	tests := map[string]string{
		"app 440":           chatbot.CApp,
		"app tf2":           chatbot.CApp,
		"game {game}":       chatbot.CApp,
		"new":               chatbot.CAppsNew,
		"players {game}":    chatbot.CAppPlayers,
		"online {game}":     chatbot.CAppPlayers,
		"popular":           chatbot.CAppsPopular,
		"random":            chatbot.CAppRandom,
		"trending":          chatbot.CAppsTrending,
		"group {game}":      chatbot.CGroup,
		"clan {game}":       chatbot.CGroup,
		"trendinggroups":    chatbot.CGroupsTrending,
		"help":              chatbot.CHelp,
		"players":           chatbot.CSteamOnline,
		"games {player}":    chatbot.CPlayerApps,
		"level {player}":    chatbot.CPlayerLevel,
		"player {player}":   chatbot.CPlayer,
		"playtime {player}": chatbot.CPlayerPlaytime,
		"recent {player}":   chatbot.CPlayerRecent,
	}

	for _, start := range []string{".", "!"} {
		for message, commandID := range tests {

			var id string

			for _, c := range chatbot.CommandRegister {
				if chatbot.RegexCache[c.Regex()].MatchString(start + message) {
					if c.ID() == commandID {
						id = commandID
						break
					}
				}
			}

			if id != commandID {
				t.Error(start+message, "!=", commandID)
			}
		}
	}
}
