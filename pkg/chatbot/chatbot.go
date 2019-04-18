package chatbot

import (
	"regexp"

	"github.com/bwmarrin/discordgo"
)

var author = &discordgo.MessageEmbedAuthor{
	Name:    "gamedb.online",
	URL:     "https://gamedb.online/",
	IconURL: "https://gamedb.online/assets/img/sa-bg-32x32.png",
}

var CommandRegister = []Command{
	CommandHelp{},
	CommandPlayer{},
	CommandLevel{},
	CommandGames{},
	CommandPlayers{},
	CommandPopular{},
	CommandRecent{},
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
