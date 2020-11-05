package chatbot

import (
	"html/template"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
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

func (CommandInvite) Description() template.HTML {
	return "Gives you the link to invite the bot to your server"
}

func (CommandInvite) Type() CommandType {
	return TypeOther
}

func (CommandInvite) Output(msg *discordgo.MessageCreate, _ steamapi.ProductCC) (message discordgo.MessageSend, err error) {

	message.Content = "See <" + config.C.DiscordBotInviteURL + ">"

	return message, nil
}
