package chatbot

import (
	"regexp"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type CommandPlayerLevel struct {
}

func (CommandPlayerLevel) Regex() *regexp.Regexp {
	return regexp.MustCompile(`^[.|!]level (.{2,32})$`)
}

func (c CommandPlayerLevel) Output(msg *discordgo.MessageCreate) (message discordgo.MessageSend, err error) {

	matches := c.Regex().FindStringSubmatch(msg.Message.Content)

	player, q, err := mongo.SearchPlayer(matches[1], bson.M{"_id": 1, "persona_name": 1, "level": 1})
	if err == mongo.ErrNoDocuments {

		message.Content = "Player **" + matches[1] + "** not found, please enter a user's vanity URL"
		return message, nil

	} else if err != nil {
		return message, err
	}

	if q {
		err = queue.ProducePlayer(queue.PlayerMessage{ID: player.ID})
		err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
		log.Err(err)
	}


	message.Content = player.GetName() + " is level **" + strconv.Itoa(player.Level) + "**"
	return message, nil
}

func (CommandPlayerLevel) Example() string {
	return ".level {player_name}"
}

func (CommandPlayerLevel) Description() string {
	return "Get the level of a player"
}

func (CommandPlayerLevel) Type() CommandType {
	return TypePlayer
}
