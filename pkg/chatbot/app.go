package chatbot

import (
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

func (CommandApp) Description() string {
	return "Retrieve information about a game"
}

func (CommandApp) Type() CommandType {
	return TypeGame
}

func (c CommandApp) LegacyInputs(input string) map[string]string {

	matches := RegexCache[c.Regex()].FindStringSubmatch(input)

	return map[string]string{
		"game": matches[2],
	}
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

func (c CommandApp) Output(_ string, region steamapi.ProductCC, inputs map[string]string) (message discordgo.MessageSend, err error) {

	apps, err := elasticsearch.SearchAppsSimple(1, inputs["game"])
	if err != nil {
		return message, err
	} else if len(apps) == 0 {
		message.Content = "Game **" + inputs["game"] + "** not found on Steam"
		return message, nil
	}

	app, err := mongo.GetApp(apps[0].ID)
	if err != nil {
		return message, err
	}

	message.Embed = getAppEmbed(c.ID(), app, region)

	return message, nil
}
