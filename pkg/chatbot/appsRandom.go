package chatbot

import (
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/olivere/elastic/v7"
)

type CommandAppRandom struct {
}

func (c CommandAppRandom) ID() string {
	return CAppsRandom
}

func (CommandAppRandom) Regex() string {
	return `^[.|!]random\s?(.*)?`
}

func (CommandAppRandom) DisableCache() bool {
	return true
}

func (CommandAppRandom) PerProdCode() bool {
	return true
}

func (CommandAppRandom) AllowDM() bool {
	return false
}

func (CommandAppRandom) Example() string {
	return ".random {tag}?"
}

func (CommandAppRandom) Description() string {
	return "Retrieve a random game, optionally by tag"
}

func (CommandAppRandom) Type() CommandType {
	return TypeGame
}

func (c CommandAppRandom) LegacyInputs(input string) map[string]string {

	matches := RegexCache[c.Regex()].FindStringSubmatch(input)

	return map[string]string{
		"tag": matches[1],
	}
}

func (c CommandAppRandom) Slash() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Name:        "tag",
			Description: "Tag",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    false,
		},
	}
}

func (c CommandAppRandom) Output(_ string, region steamapi.ProductCC, inputs map[string]string) (message discordgo.MessageSend, err error) {

	var filters = []elastic.Query{
		elastic.NewTermsQuery("type", "game", ""),
		elastic.NewRangeQuery("players").Gte(10),
	}

	if inputs["tag"] != "" {

		tag, err := mongo.GetStatByName(inputs["tag"])
		if err == mongo.ErrNoDocuments {
			message.Content = "Tag **" + inputs["tag"] + "** not found, see <" + config.C.GlobalSteamDomain + "/tags>"
			return message, nil
		} else if err != nil {
			return message, err
		}

		filters = append(filters, elastic.NewTermQuery("tags", tag.ID))
	}

	query := []elastic.Query{
		elastic.NewBoolQuery().
			Filter(filters...).
			MustNot(elastic.NewTermQuery("name", "")),
	}

	app, _, err := elasticsearch.SearchAppsRandom(query)
	if err != nil {
		return message, err
	}

	message.Embed = getAppEmbed(c.ID(), app, region)

	return message, nil
}
