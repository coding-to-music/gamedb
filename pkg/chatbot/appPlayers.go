package chatbot

import (
	"html/template"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/mongo"
)

type CommandAppPlayers struct {
}

func (c CommandAppPlayers) ID() string {
	return CAppPlayers
}

func (CommandAppPlayers) Regex() string {
	return `^[.|!](players|online) (.*)`
}

func (CommandAppPlayers) DisableCache() bool {
	return false
}

func (CommandAppPlayers) Example() string {
	return ".players {game}"
}

func (CommandAppPlayers) Description() template.HTML {
	return "Gets the number of people playing a game."
}

func (CommandAppPlayers) Type() CommandType {
	return TypeGame
}

func (c CommandAppPlayers) Output(msg *discordgo.MessageCreate) (message discordgo.MessageSend, err error) {

	matches := RegexCache[c.Regex()].FindStringSubmatch(msg.Message.Content)

	apps, _, _, err := elasticsearch.SearchApps(1, 0, matches[2], false, false, false)
	if err != nil {
		return message, err
	} else if len(apps) == 0 {
		message.Content = "Game **" + matches[2] + "** not found"
		return message, nil
	}

	app := mongo.App{}
	app.ID = apps[0].ID
	app.Name = apps[0].Name

	i, err := app.GetPlayersInGame()
	if err != nil {
		return message, err
	}

	message.Content = "<@" + msg.Author.ID + ">, " + app.GetName() + " has **" + humanize.Comma(i) + "** players"

	return message, nil
}
