package chatbot

import (
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/sql"
)

type CommandAppsNew struct {
}

func (CommandAppsNew) Regex() *regexp.Regexp {
	return regexp.MustCompile(`^[.|!]new$`)
}

func (CommandAppsNew) Output(input string) (message discordgo.MessageSend, err error) {

	message.Embed = &discordgo.MessageEmbed{
		Title:  "Popular New Apps",
		Author: author,
	}

	apps, err := sql.PopularNewApps()
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

func (CommandAppsNew) Example() string {
	return ".new"
}

func (CommandAppsNew) Description() string {
	return "Returns the most popular newly released apps"
}

func (CommandAppsNew) Type() CommandType {
	return TypeGame
}
