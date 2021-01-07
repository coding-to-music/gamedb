package chatbot

import (
	"errors"
	"html/template"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/chatbot/interactions"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/mongo"
)

type CommandApp struct {
}

func (c CommandApp) ID() string {
	return CApp
}

func (CommandApp) Regex() string {
	return `^[.|!](app|game) (.*)`
}

func (CommandApp) DisableCache() bool {
	return false
}

func (CommandApp) PerProdCode() bool {
	return true
}

func (CommandApp) Example() string {
	return ".game {game}"
}

func (CommandApp) Description() template.HTML {
	return "Get info on a game"
}

func (CommandApp) Type() CommandType {
	return TypeGame
}

func (c CommandApp) Slash() []interactions.InteractionOption {

	return []interactions.InteractionOption{
		{
			Name:        "game",
			Description: "The name or ID of the game",
			Type:        interactions.InteractionOptionTypeString,
			Required:    true,
		},
	}
}

func (c CommandApp) Output(msg *discordgo.MessageCreate, code steamapi.ProductCC) (message discordgo.MessageSend, err error) {

	matches := RegexCache[c.Regex()].FindStringSubmatch(msg.Message.Content)
	if len(matches) == 0 {
		return message, errors.New("invalid regex")
	}

	apps, err := elasticsearch.SearchAppsSimple(1, matches[2])
	if err != nil {
		return message, err
	} else if len(apps) == 0 {
		message.Content = "Game **" + matches[2] + "** not found on Steam"
		return message, nil
	}

	app, err := mongo.GetApp(apps[0].ID)
	if err != nil {
		return message, err
	}

	message.Embed = getAppEmbed(c.ID(), app, code)

	return message, nil
}
