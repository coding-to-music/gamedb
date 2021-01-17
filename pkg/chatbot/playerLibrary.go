package chatbot

import (
	"fmt"
	"strings"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/chatbot/interactions"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type CommandPlayerLibrary struct {
}

func (c CommandPlayerLibrary) ID() string {
	return CPlayerLibrary
}

func (CommandPlayerLibrary) Regex() string {
	return `^[.|!](library|top) (.{2,32})`
}

func (CommandPlayerLibrary) DisableCache() bool {
	return false
}

func (CommandPlayerLibrary) PerProdCode() bool {
	return false
}

func (CommandPlayerLibrary) Example() string {
	return ".library {player}"
}

func (CommandPlayerLibrary) Description() string {
	return "Retrieve a players top played games"
}

func (CommandPlayerLibrary) Type() CommandType {
	return TypePlayer
}

func (c CommandPlayerLibrary) LegacyInputs(input string) map[string]string {

	matches := RegexCache[c.Regex()].FindStringSubmatch(input)

	return map[string]string{
		"player": matches[2],
	}
}

func (c CommandPlayerLibrary) Slash() []interactions.InteractionOption {

	return []interactions.InteractionOption{
		{
			Name:        "player",
			Description: "The name or ID of the player",
			Type:        interactions.InteractionOptionTypeString,
			Required:    true,
		},
	}
}

func (c CommandPlayerLibrary) Output(authorID string, _ steamapi.ProductCC, inputs map[string]string) (message discordgo.MessageSend, err error) {

	player, q, err := mongo.SearchPlayer(inputs["player"], nil)
	if err == mongo.ErrNoDocuments {

		message.Content = "Player **" + inputs["player"] + "** not found, please enter a user's vanity URL"
		if q {
			message.Content += ". Player queued to be scanned."
		}
		return message, nil

	} else if err != nil {
		return message, err
	}

	if q {
		err = queue.ProducePlayer(queue.PlayerMessage{ID: player.ID}, "chatbot-player.library")
		err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
		if err != nil {
			log.ErrS(err)
		}
	}

	// Sucess response
	apps, err := mongo.GetPlayerAppsByPlayer(player.ID, 0, 10, bson.D{{"app_time", -1}}, bson.M{"app_name": 1, "app_time": 1, "app_id": 1}, nil)
	if err != nil {
		return message, err
	}

	if len(apps) > 10 {
		apps = apps[0:10]
	}

	if len(apps) > 0 {

		var code []string
		for k, app := range apps {
			code = append(code, fmt.Sprintf("%2d", k+1)+": "+app.GetTimeNice()+" - "+app.AppName)
		}

		message.Embed = &discordgo.MessageEmbed{
			Title:       player.GetName() + "'s Top Games",
			URL:         config.C.GameDBDomain + player.GetPath() + "#games",
			Author:      getAuthor(authorID),
			Color:       greenHexDec,
			Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: player.GetAvatarAbsolute(), Width: 184, Height: 184},
			Description: "```" + strings.Join(code, "\n") + "```",
		}

	} else {
		message.Content = "Profile set to private"
	}

	return message, nil
}
