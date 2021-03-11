package chatbot

import (
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/config"
)

type CommandHelp struct {
}

func (c CommandHelp) ID() string {
	return CHelp
}

func (CommandHelp) Regex() string {
	return `^[.|!]help`
}

func (CommandHelp) DisableCache() bool {
	return true
}

func (CommandHelp) PerProdCode() bool {
	return false
}

func (CommandHelp) AllowDM() bool {
	return true
}
func (CommandHelp) Example() string {
	return ".help"
}

func (CommandHelp) Description() string {
	return "Retrieve a full list of commands"
}

func (CommandHelp) Type() CommandType {
	return TypeOther
}

func (CommandHelp) LegacyInputs(_ string) map[string]string {
	return map[string]string{}
}

func (c CommandHelp) Slash() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{}
}

func (CommandHelp) Output(_ string, _ steamapi.ProductCC, _ map[string]string) (message discordgo.MessageSend, err error) {

	message.Content = "See <" + config.C.GlobalSteamDomain + "/discord-bot>"

	return message, nil
}
