package chatbot

import (
	"html/template"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/chatbot/interactions"
	"github.com/gamedb/gamedb/pkg/config"
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

func (CommandHelp) PerProdCode() bool {
	return false
}

func (CommandHelp) Example() string {
	return ".help"
}

func (CommandHelp) Description() template.HTML {
	return "Links to this list of commands"
}

func (CommandHelp) Type() CommandType {
	return TypeOther
}

func (CommandHelp) LegacyPrefix() string {
	return "help"
}

func (c CommandHelp) Slash() []interactions.InteractionOption {
	return []interactions.InteractionOption{}
}

func (CommandHelp) Output(msg *discordgo.MessageCreate, _ steamapi.ProductCC) (message discordgo.MessageSend, err error) {

	message.Content = "See " + config.C.GameDBDomain + "/discord-bot"

	return message, nil
}
