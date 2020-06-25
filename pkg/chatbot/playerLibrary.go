package chatbot

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/bwmarrin/discordgo"
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
	return `^[.|!](library|top) (.{2,32})$`
}

func (CommandPlayerLibrary) DisableCache() bool {
	return false
}

func (CommandPlayerLibrary) Example() string {
	return ".library {player}"
}

func (CommandPlayerLibrary) Description() template.HTML {
	return "Returns the players most played games"
}

func (CommandPlayerLibrary) Type() CommandType {
	return TypePlayer
}

func (c CommandPlayerLibrary) Output(msg *discordgo.MessageCreate) (message discordgo.MessageSend, err error) {

	matches := RegexCache[c.Regex()].FindStringSubmatch(msg.Message.Content)

	player, q, err := mongo.SearchPlayer(matches[2], nil)
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

	apps, err := mongo.GetPlayerAppsByPlayer(player.ID, 0, 10, bson.D{{"app_time", -1}})
	if err != nil {
		return message, err
	}

	if len(apps) > 10 {
		apps = apps[0:10]
	}

	if len(apps) > 0 {

		message.Content = "<@" + msg.Author.ID + ">"
		message.Embed = &discordgo.MessageEmbed{
			Title:  player.GetName() + "'s Top Games",
			URL:    "https://gamedb.online" + player.GetPath() + "#games",
			Author: getAuthor(msg.Author.ID),
		}

		var code []string

		for k, app := range apps {

			if k == 0 {

				avatar := app.GetIcon()
				if strings.HasPrefix(avatar, "/") {
					avatar = "https://gamedb.online" + avatar
				}

				message.Embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
					URL: avatar,
				}
			}

			code = append(code, fmt.Sprintf("%2d", k+1)+": "+app.AppName+" - "+app.GetTimeNice())
		}

		message.Embed.Description = "```" + strings.Join(code, "\n") + "```"

	} else {
		message.Content = "None" // todo, dont do as content
	}

	return message, nil
}
