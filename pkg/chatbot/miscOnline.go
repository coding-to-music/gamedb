package chatbot

import (
	"html/template"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/mongo"
)

type CommandSteamOnline struct {
}

func (c CommandSteamOnline) ID() string {
	return CSteamOnline
}

func (CommandSteamOnline) Regex() string {
	return `^[.|!](players|online)$`
}

func (CommandSteamOnline) DisableCache() bool {
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

func (c CommandSteamOnline) Output(msg *discordgo.MessageCreate) (message discordgo.MessageSend, err error) {

	var app = mongo.App{}

	i, err := app.GetPlayersOnline()
	if err != nil {
		return message, err
	}

	i2, err := app.GetPlayersInGame()
	if err != nil {
		return message, err
	}

	message.Content = "<@" + msg.Author.ID + ">, Steam has **" + humanize.Comma(i) + "** players online, **" + humanize.Comma(i2) + "** in game."

	return message, nil
}
