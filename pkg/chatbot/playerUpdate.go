package chatbot

import (
	"strconv"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/chatbot/interactions"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/oauth"
	"github.com/gamedb/gamedb/pkg/queue"
)

type CommandPlayerUpdate struct {
}

func (c CommandPlayerUpdate) ID() string {
	return CPlayerUpdate
}

func (CommandPlayerUpdate) Regex() string {
	return `^[.|!]update\s?(.{2,32})?`
}

func (CommandPlayerUpdate) DisableCache() bool {
	return true
}

func (CommandPlayerUpdate) PerProdCode() bool {
	return false
}

func (CommandPlayerUpdate) Example() string {
	return ".update {player}?"
}

func (CommandPlayerUpdate) Description() string {
	return "Updates a player's Global Steam profile"
}

func (CommandPlayerUpdate) Type() CommandType {
	return TypePlayer
}

func (c CommandPlayerUpdate) LegacyInputs(input string) map[string]string {

	matches := RegexCache[c.Regex()].FindStringSubmatch(input)

	return map[string]string{
		"player": matches[1],
	}
}

func (c CommandPlayerUpdate) Slash() []interactions.InteractionOption {

	return []interactions.InteractionOption{
		{
			Name:        "player",
			Description: "The name or ID of the player",
			Type:        interactions.InteractionOptionTypeString,
			Required:    true,
		},
	}
}

func (c CommandPlayerUpdate) Output(authorID string, _ steamapi.ProductCC, inputs map[string]string) (message discordgo.MessageSend, err error) {

	if inputs["player"] == "" {

		user, err := mysql.GetUserByProviderID(oauth.ProviderDiscord, authorID)
		if err == mysql.ErrRecordNotFound {
			message.Content = "You need to link your **Discord** account for us to know who you are: <" + config.C.GameDBDomain + "/settings>"
			return message, nil
		} else if err != nil {
			return message, err
		}

		playerID := mysql.GetUserSteamID(user.ID)
		if playerID > 0 {

			err = queue.ProducePlayer(queue.PlayerMessage{ID: playerID}, "chatbot-player.update")
			err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
			if err != nil {
				log.ErrS(err)
			}

			message.Content = "Player queued: <" + config.C.GameDBDomain + "/p" + strconv.FormatInt(playerID, 10) + ">"
		} else {
			message.Content = "You need to link your **Steam** account for us to know who you are: <" + config.C.GameDBDomain + "/settings>"
		}
		return message, nil
	}

	player, err := searchForPlayer(inputs["player"])
	if err == mongo.ErrNoDocuments {

		message.Content = "Player **" + inputs["player"] + "** not found, please enter a user's vanity URL"
		return message, nil

	} else if err != nil {
		return message, err
	}

	err = queue.ProducePlayer(queue.PlayerMessage{ID: player.ID}, "chatbot-player.update")
	err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
	if err != nil {
		log.ErrS(err)
	}

	message.Content = "Player queued: <" + config.C.GameDBDomain + "/p" + strconv.FormatInt(player.ID, 10) + ">"
	return message, nil
}
