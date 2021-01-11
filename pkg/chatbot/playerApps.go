package chatbot

import (
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

type CommandPlayerApps struct {
}

func (c CommandPlayerApps) ID() string {
	return CPlayerApps
}

func (CommandPlayerApps) Regex() string {
	return `^[.|!](games|apps) (.{2,32})`
}

func (CommandPlayerApps) DisableCache() bool {
	return false
}

func (CommandPlayerApps) PerProdCode() bool {
	return false
}

func (CommandPlayerApps) Example() string {
	return ".games {player}"
}

func (CommandPlayerApps) Description() string {
	return "Get the amount of games a player has in their library"
}

func (CommandPlayerApps) Type() CommandType {
	return TypePlayer
}

func (c CommandPlayerApps) LegacyInputs(input string) map[string]string {

	matches := RegexCache[c.Regex()].FindStringSubmatch(input)

	return map[string]string{
		"player": matches[2],
	}
}

func (c CommandPlayerApps) Slash() []interactions.InteractionOption {

	return []interactions.InteractionOption{
		{
			Name:        "player",
			Description: "The name or ID of the player",
			Type:        interactions.InteractionOptionTypeString,
			Required:    true,
		},
	}
}

func (c CommandPlayerApps) Output(_ string, _ steamapi.ProductCC, inputs map[string]string) (message discordgo.MessageSend, err error) {

	player, q, err := mongo.SearchPlayer(inputs["player"], bson.M{"_id": 1, "persona_name": 1, "games_count": 1, "ranks": 1})
	if err == mongo.ErrNoDocuments {

		message.Content = "Player **" + inputs["player"] + "** not found, please enter a user's vanity URL"
		return message, nil

	} else if err != nil {
		return message, err
	}

	if q {
		err = queue.ProducePlayer(queue.PlayerMessage{ID: player.ID}, "chatbot-player.apps")
		err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
		if err != nil {
			log.ErrS(err)
		}
	}

	var rank = "Unranked"
	if val, ok := player.Ranks[string(mongo.RankKeyGames)]; ok {
		rank = "Rank " + humanize.Comma(int64(val))
	}

	if player.GamesCount > 0 {
		message.Content = player.GetName() + " has **" + strconv.Itoa(player.GamesCount) + "** games (" + rank + ")"
	} else {
		message.Content = "Profile set to private"
	}

	return message, nil
}
