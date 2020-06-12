package chatbot

import (
	"regexp"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/elastic"
	"github.com/gamedb/gamedb/pkg/mongo"
)

type CommandAppPlayers struct {
}

func (CommandAppPlayers) Regex() *regexp.Regexp {
	// ^.(players|online) ?([a-zA-Z0-9]+)?
	return regexp.MustCompile(`^[.|!](players|online) ([a-zA-Z0-9]+)`)
}

func (CommandAppPlayers) Example() string {
	return ".players GameName"
}

func (CommandAppPlayers) Description() string {
	return "Gets the number of people playing a game."
}

func (CommandAppPlayers) Type() CommandType {
	return TypeGame
}

func (c CommandAppPlayers) Output(msg *discordgo.MessageCreate) (message discordgo.MessageSend, err error) {

	matches := RegexCache[c.Regex()].FindStringSubmatch(msg.Message.Content)

	apps, _, err := elastic.SearchApps(1, 0, matches[2], nil, false, false, false)
	if err != nil {
		return message, err
	} else if len(apps) == 0 {
		message.Content = "App **" + matches[2] + "** not found"
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
