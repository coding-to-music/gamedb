package chatbot

import (
	"regexp"

	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/website/pkg/mongo"
)

type CommandPlayer struct {
}

func (CommandPlayer) Regex() *regexp.Regexp {
	return regexp.MustCompile("^.(player|user) (.*)")
}

func (c CommandPlayer) Output(input string) (message discordgo.MessageSend, err error) {

	matches := c.Regex().FindStringSubmatch(input)

	player, err := mongo.SearchPlayer(matches[2], nil)
	if err != nil {
		return message, err
	}

	message.Content = player.GetName()

	return message, nil
}

func (CommandPlayer) Example() string {
	return ".player {playerName}"
}

func (CommandPlayer) Description() string {
	return "Get info on a player"
}
