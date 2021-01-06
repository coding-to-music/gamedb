package chatbot

import (
	"html/template"
	"strconv"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/oauth"
	"github.com/gamedb/gamedb/pkg/chatbot/interactions"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/queue"
)

type CommandPlayerUpdate struct {
}

func (c CommandPlayerUpdate) ID() string {
	return CPlayerUpdate
}

func (CommandPlayerUpdate) Regex() string {
	return `^[.|!]update\s?(.{2,32})?$`
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

func (CommandPlayerUpdate) Description() template.HTML {
	return "Updates a player, connect your discord account to leave out the a player ID"
}

func (CommandPlayerUpdate) Type() CommandType {
	return TypePlayer
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

func (c CommandPlayerUpdate) Output(msg *discordgo.MessageCreate, _ steamapi.ProductCC) (message discordgo.MessageSend, err error) {

	matches := RegexCache[c.Regex()].FindStringSubmatch(msg.Message.Content)

	if matches[1] == "" {

		user, err := mysql.GetUserByProviderID(oauth.ProviderDiscord, msg.Author.ID)
		if err == mysql.ErrRecordNotFound {
			message.Content = "You need to link your **Discord** account for us to know who you are: " + config.C.GameDBDomain + "/settings"
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

			message.Content = "Player queued: " + config.C.GameDBDomain + "/p" + strconv.FormatInt(playerID, 10)
		} else {
			message.Content = "You need to link your **Steam** account for us to know who you are: " + config.C.GameDBDomain + "/settings"
		}
		return message, nil
	}

	player, _, err := mongo.SearchPlayer(matches[1], nil)
	if err == mongo.ErrNoDocuments {

		message.Content = "Player **" + matches[1] + "** not found, please enter a user's vanity URL"
		return message, nil

	} else if err != nil {
		return message, err
	}

	err = queue.ProducePlayer(queue.PlayerMessage{ID: player.ID}, "chatbot-player.update")
	err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
	if err != nil {
		log.ErrS(err)
	}

	message.Content = "Player queued: " + config.C.GameDBDomain + "/p" + strconv.FormatInt(player.ID, 10)
	return message, nil
}
