package chatbot

import (
	"regexp"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/website/pkg/mongo"
)

type CommandGames struct {
}

func (CommandGames) Regex() *regexp.Regexp {
	return regexp.MustCompile("^.(games|apps) (.*)")
}

func (c CommandGames) Output(input string) (message discordgo.MessageSend, err error) {

	matches := c.Regex().FindStringSubmatch(input)

	player, err := mongo.SearchPlayer(matches[2], mongo.M{"_id": 1, "persona_name": 1, "games_count": 1})
	if err != nil {
		return message, err
	}

	message.Content = player.GetName() + ": " + strconv.Itoa(player.GamesCount)

	return message, nil
}

func (CommandGames) Example() string {
	return ".games {playerName}"
}

func (CommandGames) Description() string {
	return "Get the amount of games a player has in their library"
}
