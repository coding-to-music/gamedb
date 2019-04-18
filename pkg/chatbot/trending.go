package chatbot

import (
	"regexp"

	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/website/pkg/helpers"
	"github.com/gamedb/website/pkg/sql"
)

type CommandTrending struct {
}

func (c CommandTrending) Regex() *regexp.Regexp {
	return regexp.MustCompile("^.trending$")
}

func (c CommandTrending) Output(input string) (message discordgo.MessageSend, err error) {

	message.Embed = &discordgo.MessageEmbed{
		Title:  "Trending Apps",
		URL:    "https://gamedb.online/trending",
		Author: author,
	}

	apps, err := sql.TrendingApps()
	if err != nil {
		return message, err
	}

	if len(apps) >= 10 {
		for k, v := range apps[0:10] {

			message.Embed.Fields = append(message.Embed.Fields, &discordgo.MessageEmbedField{
				Name:  helpers.OrdinalComma(k + 1),
				Value: v.GetName(), // both fields are needed
			})
		}
	}

	return message, nil

}

func (c CommandTrending) Example() string {
	return ".trending"
}

func (c CommandTrending) Description() string {
	return "Returns the most positively trending apps"
}
