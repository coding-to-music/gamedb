package chatbot

import (
	"regexp"

	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/website/pkg/sql"
)

type CommandApp struct {
}

func (CommandApp) Regex() *regexp.Regexp {
	return regexp.MustCompile("^.(app|game) (.*)")
}

func (c CommandApp) Output(input string) (message discordgo.MessageSend, err error) {

	matches := c.Regex().FindStringSubmatch(input)

	app, err := sql.SearchApp(matches[2], nil)
	if err != nil {
		return message, err
	}

	message.Content = app.GetName()

	return message, nil
}

func (CommandApp) Example() string {
	return ".game {game_name}"
}

func (CommandApp) Description() string {
	return "Get info on a game"
}
