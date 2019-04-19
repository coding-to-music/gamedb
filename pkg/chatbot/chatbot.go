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
	CommandApp{},
	CommandHelp{},
	CommandPlayer{},
	CommandLevel{},
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
