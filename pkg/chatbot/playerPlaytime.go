package chatbot

import (
	"regexp"

	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type CommandPlayerPlaytime struct {
}

func (CommandPlayerPlaytime) Regex() *regexp.Regexp {
	return regexp.MustCompile(`^\.playtime (.*)`)
}

func (c CommandPlayerPlaytime) Output(input string) (message discordgo.MessageSend, err error) {

	matches := c.Regex().FindStringSubmatch(input)

	player, err := mongo.SearchPlayer(matches[1], bson.M{"_id": 1, "persona_name": 1, "play_time": 1})
	if err == mongo.ErrNoDocuments {

		message.Content = "Player **" + matches[1] + "** not found"
		return message, nil

	} else if err != nil {
		return message, err
	}

	message.Content = player.GetName() + " has played for **" + helpers.GetTimeLong(player.PlayTime, 0) + "**"
	return message, nil
}

func (CommandPlayerPlaytime) Example() string {
	return ".playtime {player_name}"
}

func (CommandPlayerPlaytime) Description() string {
	return "Get the playtime of a player"
}

func (CommandPlayerPlaytime) Type() CommandType {
	return TypePlayer
}
