package chatbot

import (
	"html/template"

	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type CommandPlayerPlaytime struct {
}

func (c CommandPlayerPlaytime) ID() string {
	return CPlayerPlaytime
}

func (CommandPlayerPlaytime) Regex() string {
	return `^[.|!]playtime (.{2,32})$`
}

func (CommandPlayerPlaytime) DisableCache() bool {
	return false
}

func (CommandPlayerPlaytime) Example() string {
	return ".playtime {player}"
}

func (CommandPlayerPlaytime) Description() template.HTML {
	return "Get the playtime of a player"
}

func (CommandPlayerPlaytime) Type() CommandType {
	return TypePlayer
}

func (c CommandPlayerPlaytime) Output(msg *discordgo.MessageCreate) (message discordgo.MessageSend, err error) {

	matches := RegexCache[c.Regex()].FindStringSubmatch(msg.Message.Content)

	player, q, err := mongo.SearchPlayer(matches[1], bson.M{"_id": 1, "persona_name": 1, "play_time": 1})
	if err == mongo.ErrNoDocuments {

		message.Content = "Player **" + matches[1] + "** not found, please enter a user's vanity URL"
		return message, nil

	} else if err != nil {
		return message, err
	}

	if q {
		err = queue.ProducePlayer(queue.PlayerMessage{ID: player.ID})
		err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
		if err != nil {
			log.ErrS(err)
		}
	}

	if player.PlayTime == 0 {
		message.Content = "<@" + msg.Author.ID + ">, Profile set to private"
	} else {
		message.Content = "<@" + msg.Author.ID + ">, " + player.GetName() + " has played for **" + helpers.GetTimeLong(player.PlayTime, 0) + "**"
	}

	return message, nil
}
