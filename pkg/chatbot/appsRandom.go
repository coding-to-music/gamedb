package chatbot

import (
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/chatbot/interactions"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/olivere/elastic/v7"
	"go.mongodb.org/mongo-driver/bson"
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

func (CommandAppRandom) Description() string {
	return "Retrieve a random game"
}

func (CommandAppRandom) Type() CommandType {
	return TypeGame
}

func (CommandAppRandom) LegacyInputs(input string) map[string]string {
	return map[string]string{}
}

func (c CommandAppRandom) Slash() []interactions.InteractionOption {
	return []interactions.InteractionOption{
		{
			Name:        "tag",
			Description: "Tag",
			Type:        interactions.InteractionOptionTypeString,
			Required:    false,
		},
	}
}

func (c CommandAppRandom) Output(_ string, region steamapi.ProductCC, inputs map[string]string) (message discordgo.MessageSend, err error) {

	var filters = []elastic.Query{
		elastic.NewTermsQuery("type", "game", ""),
		elastic.NewRangeQuery("players").Gte(10),
	}

	tag := inputs["tag"]
	if tag != "" {

		tags, err := mongo.GetStats(0, 1, bson.D{{"type", mongo.StatsTypeTags}, {"name", tag}}, nil)
		if err != nil {
			log.ErrS(err)
		} else {
			if len(tags) > 0 {
				filters = append(filters, elastic.NewTermQuery("tags", tags[0].ID))
			}
		}
	}

	query := []elastic.Query{
		elastic.NewBoolQuery().
			Filter(filters...).
			MustNot(elastic.NewTermQuery("name.raw", "")),
	}

	app, _, err := elasticsearch.SearchAppsRandom(query)
	if err != nil {
		return message, err
	}

	message.Embed = getAppEmbed(c.ID(), app, region)

	return message, nil
}
