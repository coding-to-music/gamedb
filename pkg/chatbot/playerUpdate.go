package chatbot

import (
	"strconv"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/consumers"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/oauth"
)

type CommandPlayerUpdate struct {
}

func (c CommandPlayerUpdate) ID() string {
	return CPlayerUpdate
}

func (CommandPlayerUpdate) Regex() string {
	return `^[.|!]update\s?(.+)?`
}

func (CommandPlayerUpdate) DisableCache() bool {
	return true
}

func (CommandPlayerUpdate) PerProdCode() bool {
	return false
}

func (CommandPlayerUpdate) AllowDM() bool {
	return true
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

func (c CommandPlayerUpdate) Slash() []*discordgo.ApplicationCommandOption {

	return []*discordgo.ApplicationCommandOption{
		{
			Name:        "player",
			Description: "The name or ID of the player",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    false,
		},
	}
}

func (c CommandPlayerUpdate) Output(authorID string, _ steamapi.ProductCC, inputs map[string]string) (message discordgo.MessageSend, err error) {

	if inputs["player"] == "" {

		user, err := mysql.GetUserByProviderID(oauth.ProviderDiscord, authorID)
		if err == mysql.ErrRecordNotFound {
			message.Content = "You need to link your **Discord** account for us to know who you are: <" + config.C.GlobalSteamDomain + "/settings>"
			return message, nil
		} else if err != nil {
			return message, err
		}

		playerID := mysql.GetUserSteamID(user.ID)
		if playerID > 0 {

			err = consumers.ProducePlayer(consumers.PlayerMessage{ID: playerID, ForceAchievementsRefresh: true}, "chatbot-player.update")
			err = helpers.IgnoreErrors(err, consumers.ErrInQueue)
			if err != nil {
				log.ErrS(err)
			}

			message.Content = "Player queued: <" + config.C.GlobalSteamDomain + "/p" + strconv.FormatInt(playerID, 10) + ">"
		} else {
			message.Content = "You need to link your **Steam** account for us to know who you are: <" + config.C.GlobalSteamDomain + "/settings>"
		}
		return message, nil
	}

	player, err := searchForPlayer(inputs["player"])
	if err == elasticsearch.ErrNoResult || err == steamapi.ErrProfileMissing {

		message.Content = "Player **" + inputs["player"] + "** not found, they may be set to private, please enter a user's vanity URL"
		return message, nil

	} else if err != nil {
		return message, err
	}

	err = consumers.ProducePlayer(consumers.PlayerMessage{ID: player.ID, ForceAchievementsRefresh: true}, "chatbot-player.update")
	err = helpers.IgnoreErrors(err, consumers.ErrInQueue)
	if err != nil {
		log.ErrS(err)
	}

	message.Content = "Player queued: <" + config.C.GlobalSteamDomain + "/p" + strconv.FormatInt(player.ID, 10) + ">"
	return message, nil
}
