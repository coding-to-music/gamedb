package chatbot

import (
	"regexp"

	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/config"
)

type CommandType string

var (
	TypeGame   CommandType = "game"
	TypeGames  CommandType = "games"
	TypePlayer CommandType = "player"
	TypeOther  CommandType = "Other"
)

type Command interface {
	Regex() *regexp.Regexp
	Output(input string) (discordgo.MessageSend, error)
	Example() string
	Description() string
	Type() CommandType
}

var CommandRegister = []Command{
	CommandApp{},
	CommandAppPlayers{},
	CommandAppsNew{},
	CommandAppsPopular{},
	CommandAppsTrending{},
	CommandPlayer{},
	CommandPlayerApps{},
	CommandPlayerLevel{},
	CommandPlayerPlaytime{},
	CommandPlayerRecent{},
	CommandHelp{},
}

var author = &discordgo.MessageEmbedAuthor{
	Name:    "gamedb.online",
	URL:     "https://gamedb.online/",
	IconURL: "https://gamedb.online/assets/img/sa-bg-32x32.png",
}

func init() {
	if config.IsLocal() {
		author.Name = "localhost:8081"
		author.URL = "http://localhost:8081/"
	}
}
