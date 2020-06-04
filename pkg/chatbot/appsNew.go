package chatbot

import (
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/mongo"
)

type CommandAppsNew struct {
}

func (CommandAppsNew) Regex() *regexp.Regexp {
	return regexp.MustCompile(`^[.|!]new$`)
}

func (CommandAppsNew) Example() string {
	return ".new"
}

func (CommandAppsNew) Description() string {
	return "Returns the most popular newly released games"
}

func (CommandAppsNew) Type() CommandType {
	return TypeGame
}

func (CommandAppsNew) Output(msg *discordgo.MessageCreate) (message discordgo.MessageSend, err error) {

	message.Content = "<@" + msg.Author.ID + ">"
	message.Embed = &discordgo.MessageEmbed{
		Title:  "Popular New Apps",
		Author: getAuthor(msg.Author.ID),
	}

	apps, err := mongo.PopularNewApps()
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
