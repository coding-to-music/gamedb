package chatbot

import (
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/website/pkg/helpers"
	"github.com/gamedb/website/pkg/mongo"
)

type CommandPlayerRecent struct {
}

func (CommandPlayerRecent) Regex() *regexp.Regexp {
	return regexp.MustCompile("^.recent (.*)")
}

func (c CommandPlayerRecent) Output(input string) (message discordgo.MessageSend, err error) {

	matches := c.Regex().FindStringSubmatch(input)

	player, err := mongo.SearchPlayer(matches[1], nil)
	if err != nil {
		return message, err
	}

	recent, err := player.GetRecentGames()
	if err != nil {
		return message, err
	}

	if len(recent) > 10 {
		recent = recent[0:10]
	}

	if len(recent) > 0 {

		message.Embed = &discordgo.MessageEmbed{
			Title:  "Recent Games",
			URL:    "https://gamedb.online" + player.GetPath() + "#games",
			Author: author,
		}

		var code []string

		for k, app := range recent {

			if k == 0 {
				message.Embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
					URL: helpers.GetAppIcon(app.AppID, app.ImgIconURL),
				}
			}

			space := ""
			if k < 9 && len(recent) > 9 {
				space = " "
			}

			code = append(code, "- "+space+app.Name+" - "+helpers.GetTimeShort(app.PlayTime2Weeks, 2))
		}

		message.Embed.Description = "```" + strings.Join(code, "\n") + "```"

	} else {
		message.Content = "None" // todo, dont do as content
	}

	return message, nil
}

func (CommandPlayerRecent) Example() string {
	return ".recent {player_name}"
}

func (CommandPlayerRecent) Description() string {
	return "Returns the last 10 games played by user"
}

func (CommandPlayerRecent) Type() CommandType {
	return TypePlayer
}
