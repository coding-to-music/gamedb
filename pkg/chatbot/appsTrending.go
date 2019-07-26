package chatbot

import (
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/sql"
)

type CommandAppsTrending struct {
}

func (CommandAppsTrending) Regex() *regexp.Regexp {
	return regexp.MustCompile("^.trending$")
}

func (CommandAppsTrending) Output(input string) (message discordgo.MessageSend, err error) {

	message.Embed = &discordgo.MessageEmbed{
		Title:  "Trending Apps",
		URL:    "https://gamedb.online/trending",
		Author: author,
	}

	apps, err := sql.TrendingApps()
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
				URL: "https://gamedb.online" + v.GetHeaderImage(),
			}
		}

		space := ""
		if k < 9 {
			space = " "
		}

		code = append(code, helpers.OrdinalComma(k+1)+". "+space+v.GetName()+" - "+humanize.Comma(int64(v.PlayerTrend))+" trend value")
	}

	message.Embed.Description = "```" + strings.Join(code, "\n") + "```"

	return message, nil

}

func (CommandAppsTrending) Example() string {
	return ".trending"
}

func (CommandAppsTrending) Description() string {
	return "Returns the most positively trending apps"
}

func (CommandAppsTrending) Type() CommandType {
	return TypeGames
}
