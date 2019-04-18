package chatbot

import (
	"regexp"

	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/website/pkg/helpers"
	"github.com/gamedb/website/pkg/mongo"
)

type CommandRecent struct {
}

func (c CommandRecent) Regex() *regexp.Regexp {
	return regexp.MustCompile("^.recent (.*)")
}

func (c CommandRecent) Output(input string) (message discordgo.MessageSend, err error) {

	matches := c.Regex().FindStringSubmatch(input)

	player, err := mongo.SearchPlayer(matches[1], nil)
	if err != nil {
		return message, err
	}

	message.Embed = &discordgo.MessageEmbed{
		Title:  "Recent Games",
		URL:    "https://gamedb.online" + player.GetPath() + "#games",
		Author: author,
	}

	recent, err := player.GetRecentGames()
	if err != nil {
		return message, err
	}

	if len(recent) > 0 {
		if len(recent) > 10 {
			recent = recent[0:10]
		}

		for k, v := range recent {

			message.Embed.Fields = append(message.Embed.Fields, &discordgo.MessageEmbedField{
				Name:  helpers.OrdinalComma(k + 1),
				Value: v.Name,
			})
		}
	} else {
		message.Content = "None" // todo, dont do as content
	}

	return message, nil
}

func (c CommandRecent) Example() string {
	return ".recent {playerName}"
}

func (c CommandRecent) Description() string {
	return "Returns the last 10 games played by user"
}
