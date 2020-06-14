package chatbot

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/mongo"
)

type CommandAppsTrending struct {
}

func (CommandAppsTrending) Regex() string {
	return `^[.|!]trending$`
}

func (CommandAppsTrending) DisableCache() bool {
	return false
}

func (CommandAppsTrending) Example() string {
	return ".trending"
}

func (CommandAppsTrending) Description() string {
	return "Returns the top trending games"
}

func (CommandAppsTrending) Type() CommandType {
	return TypeGame
}

func (CommandAppsTrending) Output(msg *discordgo.MessageCreate) (message discordgo.MessageSend, err error) {

	message.Content = "<@" + msg.Author.ID + ">"
	message.Embed = &discordgo.MessageEmbed{
		Title:  "Trending Apps",
		URL:    "https://gamedb.online/trending",
		Author: getAuthor(msg.Author.ID),
	}

	apps, err := mongo.TrendingApps()
	if err != nil {
		return message, err
	}

	if len(apps) > 10 {
		apps = apps[0:10]
	}

	var code []string

	for k, app := range apps {

		avatar := app.GetHeaderImage()
		if strings.HasPrefix(avatar, "/") {
			avatar = "https://gamedb.online" + avatar
		}

		if k == 0 {
			message.Embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
				URL: avatar,
			}
		}

		space := ""
		if k < 9 {
			space = " "
		}

		code = append(code, helpers.OrdinalComma(k+1)+". "+space+app.GetName()+" - "+humanize.Comma(app.PlayerTrend)+" trend value")
	}

	message.Embed.Description = "```" + strings.Join(code, "\n") + "```"

	return message, nil
}
