package chatbot

import (
	"regexp"

	"github.com/bwmarrin/discordgo"
)

type CommandTrending struct {
}

func (c CommandTrending) Regex() *regexp.Regexp {
	return regexp.MustCompile("^.trending$")
}

func (c CommandTrending) Output(input string) (message discordgo.MessageSend, err error) {
	panic("implement me")
}

func (c CommandTrending) Example() string {
	panic("implement me")
}

func (c CommandTrending) Description() string {
	panic("implement me")
}
