package chatbot

import (
	"github.com/bwmarrin/discordgo"
)

type CommandHelp struct {
}

func (c CommandHelp) ID() string {
	return CHelp
}

func (CommandHelp) Regex() string {
	return `^[.|!]help$`
}

func (CommandHelp) DisableCache() bool {
	return true
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

func (CommandHelp) Output(msg *discordgo.MessageCreate) (message discordgo.MessageSend, err error) {

	message.Content = "See https://gamedb.online/discord-bot"

	return message, nil
}
