package chatbot

import (
	"regexp"

	"github.com/bwmarrin/discordgo"
)

var CommandRegister = []Command{
	CommandHelp{},
	CommandPlayer{},
	CommandPlayers{},
	CommandPopular{},
	CommandRecent{},
	CommandTrending{},
}

type Command interface {
	Regex() *regexp.Regexp
	Output(input string) (discordgo.MessageSend, error)
	Example() string
	Description() string
}

// .game 123 |.app half life
// .user 123 |.user jimeagle
// .recent 123|jimeagle
// .trending - top 10
// .popular - top 10 based on players
