package chatbot

import (
	"errors"
	"html/template"

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

type CommandPlayerPlaytime struct {
}

func (c CommandPlayerPlaytime) ID() string {
	return CPlayerPlaytime
}

func (CommandPlayerPlaytime) Regex() string {
	return `^[.|!]playtime (.{2,32})`
}

func (CommandPlayerPlaytime) DisableCache() bool {
	return false
}

func (CommandPlayerPlaytime) PerProdCode() bool {
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

func (CommandPlayerPlaytime) LegacyPrefix() string {
	return "playtime"
}

func (c CommandPlayerPlaytime) Slash() []interactions.InteractionOption {

	return []interactions.InteractionOption{
		{
			Name:        "player",
			Description: "The name or ID of the player",
			Type:        interactions.InteractionOptionTypeString,
			Required:    true,
		},
	}
}

func (c CommandPlayerPlaytime) Output(msg *discordgo.MessageCreate, _ steamapi.ProductCC) (message discordgo.MessageSend, err error) {

	matches := RegexCache[c.Regex()].FindStringSubmatch(msg.Message.Content)
	if len(matches) == 0 {
		return message, errors.New("invalid regex")
	}

	player, q, err := mongo.SearchPlayer(matches[1], bson.M{"_id": 1, "persona_name": 1, "play_time": 1, "ranks": 1})
	if err == mongo.ErrNoDocuments {

		message.Content = "Player **" + matches[1] + "** not found, please enter a user's vanity URL"
		return message, nil

	} else if err != nil {
		return message, err
	}

	if q {
		err = queue.ProducePlayer(queue.PlayerMessage{ID: player.ID}, "chatbot-player.playtime")
		err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
		if err != nil {
			log.ErrS(err)
		}
	}

	var rank = "Unranked"
	if val, ok := player.Ranks[string(mongo.RankKeyPlaytime)]; ok {
		rank = "Rank " + humanize.Comma(int64(val))
	}

	if player.PlayTime == 0 {
		message.Content = "Profile set to private"
	} else {
		message.Content = player.GetName() + " has played for **" + helpers.GetTimeLong(player.PlayTime, 0) + "**" +
			" (" + rank + ")"
	}

	return message, nil
}
