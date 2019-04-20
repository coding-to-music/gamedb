package chatbot

import (
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type CommandPlayers struct {
}

func (CommandPlayers) Regex() *regexp.Regexp {
	return regexp.MustCompile("^.players [a-zA-Z0-9]+")
}

func (CommandPlayers) Output(input string) (message discordgo.MessageSend, err error) {

	input = strings.TrimPrefix(input, ".players ")

	message.Content = ""

	return message, nil
}

func (CommandPlayers) Example() string {
	return ".players {playerName}"
}

func (CommandPlayers) Description() string {
	return "Gets the number of people playing."
}
