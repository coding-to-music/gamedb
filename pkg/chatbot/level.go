package chatbot

import (
	"regexp"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/website/pkg/mongo"
)

type CommandLevel struct {
}

func (CommandLevel) Regex() *regexp.Regexp {
	return regexp.MustCompile("^.level (.*)")
}

func (c CommandLevel) Output(input string) (message discordgo.MessageSend, err error) {

	matches := c.Regex().FindStringSubmatch(input)

	player, err := mongo.SearchPlayer(matches[1], mongo.M{"_id": 1, "persona_name": 1, "level": 1})
	if err != nil {
		return message, err
	}

	message.Content = player.GetName() + ": " + strconv.Itoa(player.Level)

	return message, nil
}

func (CommandLevel) Example() string {
	return ".level {playerName}"
}

func (CommandLevel) Description() string {
	return "Get the level of a player"
}
