package chatbot

import (
	"errors"
	"html/template"
	"strconv"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/chatbot/interactions"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type CommandPlayerLevel struct {
}

func (c CommandPlayerLevel) ID() string {
	return CPlayerLevel
}

func (CommandPlayerLevel) Regex() string {
	return `^[.|!]level (.{2,32})$`
}

func (CommandPlayerLevel) DisableCache() bool {
	return false
}

func (CommandPlayerLevel) PerProdCode() bool {
	return false
}

func (CommandPlayerLevel) Example() string {
	return ".level {player}"
}

func (CommandPlayerLevel) Description() template.HTML {
	return "Get the level of a player"
}

func (CommandPlayerLevel) Type() CommandType {
	return TypePlayer
}

func (c CommandPlayerLevel) Slash() []interactions.InteractionOption {

	return []interactions.InteractionOption{
		{
			Name:        "player",
			Description: "The name or ID of the player",
			Type:        interactions.InteractionOptionTypeString,
			Required:    true,
		},
	}
}

func (c CommandPlayerLevel) Output(msg *discordgo.MessageCreate, _ steamapi.ProductCC) (message discordgo.MessageSend, err error) {

	matches := RegexCache[c.Regex()].FindStringSubmatch(msg.Message.Content)
	if len(matches) == 0 {
		return message, errors.New("invalid regex")
	}

	player, q, err := mongo.SearchPlayer(matches[1], bson.M{"_id": 1, "persona_name": 1, "level": 1, "ranks": 1})
	if err == mongo.ErrNoDocuments {

		message.Content = "Player **" + matches[1] + "** not found, please enter a user's vanity URL"
		return message, nil

	} else if err != nil {
		return message, err
	}

	if q {
		err = queue.ProducePlayer(queue.PlayerMessage{ID: player.ID}, "chatbot-player.level")
		err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
		if err != nil {
			log.ErrS(err)
		}
	}

	var rank = "Unranked"
	if val, ok := player.Ranks[string(mongo.RankKeyLevel)]; ok {
		rank = "Rank " + humanize.Comma(int64(val))
	}

	message.Content = player.GetName() + " is level **" + strconv.Itoa(player.Level) + "**" +
		" (" + rank + ")"
	return message, nil
}
