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

type CommandPlayerApps struct {
}

func (CommandPlayerApps) Regex() *regexp.Regexp {
	return regexp.MustCompile(`^[.|!](games|apps) (.{2,32})$`)
}

func (c CommandPlayerApps) Output(msg *discordgo.MessageCreate) (message discordgo.MessageSend, err error) {

	matches := c.Regex().FindStringSubmatch(msg.Message.Content)

	player, q, err := mongo.SearchPlayer(matches[2], bson.M{"_id": 1, "persona_name": 1, "games_count": 1})
	if err == mongo.ErrNoDocuments {

		message.Content = "Player **" + matches[2] + "** not found, please enter a user's vanity URL"
		return message, nil

	} else if err != nil {
		return message, err
	}

	if q {
		err = queue.ProducePlayer(queue.PlayerMessage{ID: player.ID})
		err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
		log.Err(err)
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
