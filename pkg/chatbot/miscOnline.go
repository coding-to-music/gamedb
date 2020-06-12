package chatbot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/mongo"
)

type CommandAppPlayersSteam struct {
}

func (CommandAppPlayersSteam) Regex() string {
	return `^[.|!](players|online)$`
}

func (CommandAppPlayersSteam) DisableCache() bool {
	return false
}

func (CommandAppPlayersSteam) Example() string {
	return ".players"
}

func (CommandAppPlayersSteam) Description() string {
	return "Gets the number of people on Steam."
}

func (CommandAppPlayersSteam) Type() CommandType {
	return TypeOther
}

func (c CommandAppPlayersSteam) Output(msg *discordgo.MessageCreate) (message discordgo.MessageSend, err error) {

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
