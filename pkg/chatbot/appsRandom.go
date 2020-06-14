package chatbot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type CommandAppRandom struct {
}

func (CommandAppRandom) Regex() string {
	return `^[.|!]random$`
}

func (CommandAppRandom) DisableCache() bool {
	return true
}

func (CommandAppRandom) Example() string {
	return ".random"
}

func (CommandAppRandom) Description() string {
	return "Get a random game"
}

func (CommandAppRandom) Type() CommandType {
	return TypeGame
}

func (c CommandAppRandom) Output(msg *discordgo.MessageCreate) (message discordgo.MessageSend, err error) {

	var filter = bson.D{
		{"$or", bson.A{
			bson.M{"type": "game"},
			bson.M{"type": ""},
		}},
		{"name", bson.M{"$ne": ""}},
	}

	var projection = bson.M{"_id": 1, "name": 1, "prices": 1, "release_date": 1, "release_date_unix": 1, "reviews_score": 1, "group_id": 1, "group_followers": 1}

	apps, err := mongo.GetRandomApps(1, filter, projection)
	if err != nil {
		return message, err
	}

	if len(apps) > 0 {

		var app = apps[0]

		message.Content = "<@" + msg.Author.ID + ">"
		message.Embed = getAppEmbed(app)
	}

	return message, nil
}
