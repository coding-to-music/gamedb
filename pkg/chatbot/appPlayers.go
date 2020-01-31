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

func (c CommandAppPlayers) Output(msg *discordgo.MessageCreate) (message discordgo.MessageSend, err error) {

	matches := c.Regex().FindStringSubmatch(msg.Message.Content)

	app, err := mongo.SearchApps(matches[2], nil)
	if err == mongo.ErrNoDocuments || err == mongo.ErrInvalidAppID {

		message.Content = "App **" + matches[1] + "** not found"
		return message, nil

	} else if err != nil {
		return message, err
	}

	i, err := app.GetOnlinePlayers()
	if err != nil {
		return message, err
	}

	message.Content = app.GetName() + " has **" + humanize.Comma(int64(i)) + "** players"

	return message, nil
}

func (CommandAppPlayers) Example() string {
	return ".players {app_name}"
}

func (CommandAppPlayers) Description() string {
	return "Gets the number of people playing."
}

func (CommandAppPlayers) Type() CommandType {
	return TypeGame
}
