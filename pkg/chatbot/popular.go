package chatbot

import (
	"regexp"

	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/website/pkg/helpers"
	"github.com/gamedb/website/pkg/sql"
)

type CommandPopular struct {
}

func (c CommandPopular) Regex() *regexp.Regexp {
	return regexp.MustCompile("^.popular$")
}

func (c CommandPopular) Output(input string) (message discordgo.MessageSend, err error) {

	message.Embed = &discordgo.MessageEmbed{
		Title:  "Popular Apps",
		Author: author,
	}

	apps, err := sql.PopularApps()
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

func (c CommandPopular) Example() string {
	return ".popular"
}

func (c CommandPopular) Description() string {
	return "Returns the most popular apps in order of players over the last week"
}
