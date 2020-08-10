package chatbot

import (
	"html/template"
	"strconv"

	"github.com/bwmarrin/discordgo"
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

func (CommandPlayerUpdate) Example() string {
	return ".update {player}?"
}

func (CommandPlayerUpdate) Description() template.HTML {
	return "Updates a player, connect your discord account to leave out the a player ID"
}

func (CommandPlayerUpdate) Type() CommandType {
	return TypePlayer
}

func (c CommandPlayerUpdate) Output(msg *discordgo.MessageCreate) (message discordgo.MessageSend, err error) {

	matches := RegexCache[c.Regex()].FindStringSubmatch(msg.Message.Content)

	if matches[1] == "" {

		user, err := mysql.GetUserByKey("discord_id", msg.Author.ID, 0)
		if err == mysql.ErrRecordNotFound {
			message.Content = "You need to link your **Discord** account for us to know who you are: https://gamedb.online/settings"
			return message, nil
		} else if err != nil {
			return message, err
		}

		playerID := user.GetSteamID()
		if playerID > 0 {

			err = queue.ProducePlayer(queue.PlayerMessage{ID: playerID})
			err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
			log.Err(err)

			message.Content = "Player queued: https://gamedb.online/p" + user.SteamID.String
		} else {
			message.Content = "You need to link your **Steam** account for us to know who you are: https://gamedb.online/settings"
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

	err = queue.ProducePlayer(queue.PlayerMessage{ID: player.ID})
	err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
	log.Err(err)

	message.Content = "Player queued: https://gamedb.online/p" + strconv.FormatInt(player.ID, 10)
	return message, nil
}
