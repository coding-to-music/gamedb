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
		Title: "Popular Apps",
		URL:   "https://gamedb.online/apps/440",
		// Fields: [],
		Author: &discordgo.MessageEmbedAuthor{
			Name:    "gamedb.online",
			URL:     "https://gamedb.online/",
			IconURL: "https://gamedb.online/assets/img/sa-bg-32x32.png",
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "x",
			IconURL: "https://gamedb.online/assets/img/sa-bg-32x32.png",
		},
	}

	apps, err := sql.PopularApps()
	if err != nil {
		return message, err
	}

	if len(apps) >= 10 {
		for k, v := range apps[0:10] {

			message.Embed.Fields = append(message.Embed.Fields, &discordgo.MessageEmbedField{
				Name:  helpers.OrdinalComma(k),
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
