package chatbot

import (
	"html/template"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/chatbot/interactions"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/olivere/elastic/v7"
)

type CommandAppRandom struct {
}

func (c CommandAppRandom) ID() string {
	return CAppsRandom
}

func (CommandAppRandom) Regex() string {
	return `^[.|!]random`
}

func (CommandAppRandom) DisableCache() bool {
	return true
}

func (CommandAppRandom) PerProdCode() bool {
	return true
}

func (CommandAppRandom) Example() string {
	return ".random"
}

func (CommandAppRandom) Description() template.HTML {
	return "Get a random game"
}

func (CommandAppRandom) Type() CommandType {
	return TypeGame
}

func (CommandAppRandom) LegacyPrefix() string {
	return "random$"
}

func (c CommandAppRandom) Slash() []interactions.InteractionOption {
	return []interactions.InteractionOption{}
}

func (c CommandAppRandom) Output(msg *discordgo.MessageCreate, code steamapi.ProductCC) (message discordgo.MessageSend, err error) {

	filters := []elastic.Query{
		elastic.NewBoolQuery().
			Filter(
				elastic.NewTermsQuery("type", "game", ""),
			).
			MustNot(
				elastic.NewTermQuery("name.raw", ""),
			),
	}

	app, _, err := elasticsearch.SearchAppsRandom(filters)
	if err != nil {
		return message, err
	}

	message.Embed = getAppEmbed(c.ID(), app, code)

	return message, nil
}
