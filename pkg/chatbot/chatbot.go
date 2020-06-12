package chatbot

import (
	"regexp"

	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/config"
)

type CommandType string

var (
	TypeGame   CommandType = "Game"
	TypePlayer CommandType = "Player"
	TypeGroup  CommandType = "Group"
	TypeOther  CommandType = "Miscellaneous"

	RegexCache = make(map[string]*regexp.Regexp, len(CommandRegister))
)

type Command interface {
	Regex() string
	DisableCache() bool
	Output(*discordgo.MessageCreate) (discordgo.MessageSend, error)
	Example() string
	Description() string
	Type() CommandType
}

var CommandRegister = []Command{
	CommandApp{},
	CommandAppPlayers{},
	CommandAppPlayersSteam{},
	CommandAppRandom{},
	CommandAppsNew{},
	CommandAppsPopular{},
	CommandAppsTrending{},
	CommandGroup{},
	CommandPlayer{},
	CommandPlayerApps{},
	CommandPlayerLevel{},
	CommandPlayerPlaytime{},
	CommandPlayerRecent{},
	CommandHelp{},
}

func getAuthor(guildID string) *discordgo.MessageEmbedAuthor {

	author := &discordgo.MessageEmbedAuthor{
		Name:    "gamedb.online",
		URL:     "https://gamedb.online/?utm_source=discord&utm_medium=discord&utm_content=" + guildID,
		IconURL: "https://gamedb.online/assets/img/sa-bg-32x32.png",
	}
	if config.IsLocal() {
		author.Name = "localhost:" + config.Config.WebserverPort.Get()
		author.URL = "http://localhost:" + config.Config.WebserverPort.Get() + "/"
	}
	return author
}

func getFooter() *discordgo.MessageEmbedFooter {

	footer := &discordgo.MessageEmbedFooter{
		Text:         "gamedb.online",
		IconURL:      "https://gamedb.online/assets/img/sa-bg-32x32.png",
		ProxyIconURL: "",
	}

	if config.IsLocal() {
		footer.Text = "localhost:" + config.Config.WebserverPort.Get()
	}

	return footer
}
