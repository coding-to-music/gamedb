package chatbot

import (
	"regexp"

	"github.com/bwmarrin/discordgo"
)

type CommandRecent struct {
}

func (c CommandRecent) Regex() *regexp.Regexp {
	return regexp.MustCompile("^.recent (.*)")
}

func (c CommandRecent) Output(input string) (message discordgo.MessageSend, err error) {
	return message, nil // todo
}

func (c CommandRecent) Example() string {
	return ".recent username"
}

func (c CommandRecent) Description() string {
	return "Returns the last 10 games played by user"
}
