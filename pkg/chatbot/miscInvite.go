package chatbot

import (
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/chatbot/interactions"
	"github.com/gamedb/gamedb/pkg/config"
)

type CommandInvite struct {
}

func (c CommandInvite) ID() string {
	return CInvite
}

func (CommandInvite) Regex() string {
	return `^[.|!]invite$`
}

func (CommandInvite) DisableCache() bool {
	return true
}

func (CommandInvite) PerProdCode() bool {
	return false
}

func (CommandInvite) Example() string {
	return ".invite"
}

func (CommandInvite) Description() string {
	return "Retrieve a bot invite link for your Discord server"
}

func (CommandInvite) Type() CommandType {
	return TypeOther
}

func (CommandInvite) LegacyInputs(_ string) map[string]string {
	return map[string]string{}
}

func (c CommandInvite) Slash() []interactions.InteractionOption {
	return []interactions.InteractionOption{}
}

func (CommandInvite) Output(_ string, _ steamapi.ProductCC, _ map[string]string) (message discordgo.MessageSend, err error) {

	message.Content = "See <" + config.C.DiscordBotInviteURL + ">"

	return message, nil
}
