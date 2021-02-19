package chatbot

import (
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/chatbot/interactions"
)

type CommandFeedback struct {
}

func (c CommandFeedback) ID() string {
	return CFeedback
}

func (CommandFeedback) Regex() string {
	return `^[.|!](feedback|support)$`
}

func (CommandFeedback) DisableCache() bool {
	return true
}

func (CommandFeedback) PerProdCode() bool {
	return false
}

func (CommandFeedback) Example() string {
	return ".feedback"
}

func (CommandFeedback) Description() string {
	return "Retrieve a link to send feedback"
}

func (CommandFeedback) Type() CommandType {
	return TypeOther
}

func (CommandFeedback) LegacyInputs(_ string) map[string]string {
	return map[string]string{}
}

func (c CommandFeedback) Slash() []interactions.InteractionOption {
	return []interactions.InteractionOption{}
}

func (CommandFeedback) Output(_ string, _ steamapi.ProductCC, _ map[string]string) (message discordgo.MessageSend, err error) {

	message.Content = "https://discord.gg/c5zrcus"

	return message, nil
}
