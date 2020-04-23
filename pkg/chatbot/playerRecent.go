package chatbot

import (
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type CommandPlayerRecent struct {
}

func (CommandPlayerRecent) Regex() *regexp.Regexp {
	return regexp.MustCompile(`^[.|!]recent (.{2,32})$`)
}

func (CommandPlayerRecent) Example() string {
	return ".recent PlayerName"
}

func (CommandPlayerRecent) Description() string {
	return "Returns the last 10 games played by user"
}

func (CommandPlayerRecent) Type() CommandType {
	return TypePlayer
}

func (c CommandPlayerRecent) Output(msg *discordgo.MessageCreate) (message discordgo.MessageSend, err error) {

	matches := c.Regex().FindStringSubmatch(msg.Message.Content)

	player, q, err := mongo.SearchPlayer(matches[1], nil)
	if err == mongo.ErrNoDocuments {

		message.Content = "Player **" + matches[1] + "** not found, please enter a user's vanity URL"
		return message, nil

	} else if err != nil {
		return message, err
	}

	if q {
		err = queue.ProducePlayer(queue.PlayerMessage{ID: player.ID})
		err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
		log.Err(err)
	}

	recent, err := mongo.GetRecentApps(player.ID, 0, 10, bson.D{{"playtime_2_weeks", -1}})
	if err != nil {
		return message, err
	}

	if len(recent) > 10 {
		recent = recent[0:10]
	}

	if len(recent) > 0 {

		message.Content = "<@" + msg.Author.ID + ">"
		message.Embed = &discordgo.MessageEmbed{
			Title:  "Recent Games",
			URL:    "https://gamedb.online" + player.GetPath() + "#games",
			Author: getAuthor(msg.Author.ID),
		}

		var code []string

		for k, app := range recent {

			avatar := helpers.GetAppIcon(app.AppID, app.Icon)
			if strings.HasPrefix(avatar, "/") {
				avatar = "https://gamedb.online" + avatar
			}

			if k == 0 {
				message.Embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
					URL: avatar,
				}
			}

			space := ""
			if k < 9 && len(recent) > 9 {
				space = " "
			}

			code = append(code, "- "+space+app.AppName+" - "+helpers.GetTimeShort(app.PlayTime2Weeks, 2))
		}

		message.Embed.Description = "```" + strings.Join(code, "\n") + "```"

	} else {
		message.Content = "None" // todo, dont do as content
	}

	return message, nil
}
