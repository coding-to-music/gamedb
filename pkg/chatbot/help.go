package chatbot

import (
	"regexp"

	"github.com/bwmarrin/discordgo"
)

type CommandHelp struct {
}

func (CommandHelp) Regex() *regexp.Regexp {
	return regexp.MustCompile(`^[.|!]help$`)
}

func (CommandHelp) Output(msg *discordgo.MessageCreate) (message discordgo.MessageSend, err error) {

	message.Content = "See https://gamedb.online/chat-bot"

	return message, nil
}

func (CommandHelp) Example() string {
	return ".help"
}

func (CommandHelp) Description() string {
	return "Links to this list of commands"
}

func (CommandHelp) Type() CommandType {
	return TypeOther
}
