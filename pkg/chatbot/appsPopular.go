package chatbot

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/mongo"
)

type CommandAppsPopular struct {
}

func (c CommandAppsPopular) ID() string {
	return CAppsPopular
}

func (CommandAppsPopular) Regex() string {
	return `^[.|!]popular$`
}

func (CommandAppsPopular) DisableCache() bool {
	return false
}

func (CommandAppsPopular) Example() string {
	return ".popular"
}

func (CommandAppsPopular) Description() string {
	return "Returns the most popular games in order of players over the last week"
}

func (CommandAppsPopular) Type() CommandType {
	return TypeGame
}

func (CommandAppsPopular) Output(msg *discordgo.MessageCreate) (message discordgo.MessageSend, err error) {

	message.Content = "<@" + msg.Author.ID + ">"
	message.Embed = &discordgo.MessageEmbed{
		Title:  "Popular Apps",
		Author: getAuthor(msg.Author.ID),
	}

	apps, err := mongo.PopularApps()
	if err != nil {
		return message, err
	}

	if len(apps) > 10 {
		apps = apps[0:10]
	}

	var code []string

	for k, v := range apps {

		if k == 0 {
			message.Embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
				URL: v.GetHeaderImage(),
			}
		}

		space := ""
		if k < 9 {
			space = " "
		}

		code = append(code, helpers.OrdinalComma(k+1)+". "+space+v.GetName()+" - "+humanize.Comma(int64(v.PlayerPeakWeek))+" players")
	}

	message.Embed.Description = "```" + strings.Join(code, "\n") + "```"

	return message, nil
}
