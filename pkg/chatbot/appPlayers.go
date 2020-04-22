package chatbot

import (
	"regexp"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/mongo"
)

type CommandAppPlayers struct {
}

func (CommandAppPlayers) Regex() *regexp.Regexp {
	// ^.(players|online) ?([a-zA-Z0-9]+)?
	return regexp.MustCompile(`^[.|!](players|online) ([a-zA-Z0-9]+)`)
}

func (CommandAppPlayers) Example() string {
	return ".players {app_name}"
}

func (CommandAppPlayers) Description() string {
	return "Gets the number of people playing a game."
}

func (CommandAppPlayers) Type() CommandType {
	return TypeGame
}

func (c CommandAppPlayers) Output(msg *discordgo.MessageCreate) (message discordgo.MessageSend, err error) {

	matches := c.Regex().FindStringSubmatch(msg.Message.Content)

	app, err := mongo.SearchApps(matches[2], nil)
	if err == mongo.ErrNoDocuments || err == mongo.ErrInvalidAppID {

		message.Content = "App **" + matches[2] + "** not found"
		return message, nil

	} else if err != nil {
		return message, err
	}

	i, err := app.GetPlayersInGame()
	if err != nil {
		return message, err
	}

	message.Content = "<@" + msg.Author.ID + ">, " + app.GetName() + " has **" + humanize.Comma(i) + "** players"

	return message, nil
}
