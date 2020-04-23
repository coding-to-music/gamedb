package chatbot

import (
	"regexp"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type CommandApp struct {
}

func (CommandApp) Regex() *regexp.Regexp {
	return regexp.MustCompile(`^[.|!](app|game) (.*)`)
}

func (CommandApp) Example() string {
	return ".game GameName"
}

func (CommandApp) Description() string {
	return "Get info on a game"
}

func (CommandApp) Type() CommandType {
	return TypeGame
}

func (c CommandApp) Output(msg *discordgo.MessageCreate) (message discordgo.MessageSend, err error) {

	matches := c.Regex().FindStringSubmatch(msg.Message.Content)

	app, err := mongo.SearchApps(matches[2], bson.M{"_id": 1, "name": 1, "prices": 1, "release_date": 1, "release_date_unix": 1, "reviews_score": 1, "group_followers": 1})
	if err == mongo.ErrNoDocuments || err == mongo.ErrInvalidAppID {

		message.Content = "App **" + matches[2] + "** not found"
		return message, nil

	} else if err != nil {
		return message, err
	}

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

	return message, nil
}
