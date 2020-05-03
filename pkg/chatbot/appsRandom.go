package chatbot

import (
	"regexp"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type CommandAppRandom struct {
}

func (CommandAppRandom) Regex() *regexp.Regexp {
	return regexp.MustCompile(`^[.|!]random$`)
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
		message.Embed = &discordgo.MessageEmbed{
			Title: app.GetName(),
			URL:   "https://gamedb.online" + app.GetPath(),
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: app.GetHeaderImage(),
			},
			Footer: getFooter(),
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:  "Release Date",
					Value: app.GetReleaseDateNice(),
				},
				{
					Name:  "Price",
					Value: app.Prices.Get(steamapi.ProductCCUS).GetFinal(),
				},
				{
					Name:  "Review Score",
					Value: app.GetReviewScore(),
				},
				{
					Name:  "Followers",
					Value: app.GetFollowers(),
				},
			},
		}

	}

	return message, nil
}
