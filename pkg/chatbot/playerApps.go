package chatbot

import (
	"regexp"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type CommandPlayerApps struct {
}

func (CommandPlayerApps) Regex() *regexp.Regexp {
	return regexp.MustCompile(`^\.(games|apps) (.*)`)
}

func (c CommandPlayerApps) Output(input string) (message discordgo.MessageSend, err error) {

	matches := c.Regex().FindStringSubmatch(input)

	player, err := mongo.SearchPlayer(matches[2], bson.M{"_id": 1, "persona_name": 1, "games_count": 1})
	if err == mongo.ErrNoDocuments {

		message.Content = "Player **" + matches[2] + "** not found"
		return message, nil

	} else if err != nil {
		return message, err
	}

	message.Content = player.GetName() + " has **" + strconv.Itoa(player.GamesCount) + "** apps"
	return message, nil
}

func (CommandPlayerApps) Example() string {
	return ".games {player_name}"
}

func (CommandPlayerApps) Description() string {
	return "Get the amount of games a player has in their library"
}

func (CommandPlayerApps) Type() CommandType {
	return TypePlayer
}
