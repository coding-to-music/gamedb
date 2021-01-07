package chatbot

import (
	"html/template"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/chatbot/interactions"
	"github.com/gamedb/gamedb/pkg/mongo"
)

type CommandSteamOnline struct {
}

func (c CommandSteamOnline) ID() string {
	return CSteamOnline
}

func (CommandSteamOnline) Regex() string {
	return `^[.|!](players|online)`
}

func (CommandSteamOnline) DisableCache() bool {
	return false
}

func (CommandSteamOnline) PerProdCode() bool {
	return false
}

func (CommandSteamOnline) Example() string {
	return ".players"
}

func (CommandSteamOnline) Description() template.HTML {
	return "Gets the number of people on Steam."
}

func (CommandSteamOnline) Type() CommandType {
	return TypeOther
}

func (CommandSteamOnline) LegacyPrefix() string {
	return "players"
}

func (c CommandSteamOnline) Slash() []interactions.InteractionOption {

	return []interactions.InteractionOption{}
}

func (c CommandSteamOnline) Output(msg *discordgo.MessageCreate, _ steamapi.ProductCC) (message discordgo.MessageSend, err error) {

	var app = mongo.App{}

	i, err := app.GetPlayersOnline()
	if err != nil {
		return message, err
	}

	i2, err := app.GetPlayersInGame()
	if err != nil {
		return message, err
	}

	message.Content = "Steam has **" + humanize.Comma(i) + "** players online, **" + humanize.Comma(i2) + "** in game."

	return message, nil
}
