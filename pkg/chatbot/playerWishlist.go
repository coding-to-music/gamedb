package chatbot

import (
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type CommandPlayerWishlist struct {
}

func (c CommandPlayerWishlist) ID() string {
	return CPlayerWishlist
}

func (CommandPlayerWishlist) Regex() string {
	return `^[.|!]wishlist (.{2,32})$`
}

func (CommandPlayerWishlist) DisableCache() bool {
	return false
}

func (CommandPlayerWishlist) Example() string {
	return ".wishlist {player}"
}

func (CommandPlayerWishlist) Description() string {
	return "Lists a player's wishlist"
}

func (CommandPlayerWishlist) Type() CommandType {
	return TypePlayer
}

func (c CommandPlayerWishlist) Output(msg *discordgo.MessageCreate) (message discordgo.MessageSend, err error) {

	matches := RegexCache[c.Regex()].FindStringSubmatch(msg.Message.Content)

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

	wishlistApps, err := mongo.GetPlayerWishlistAppsByPlayer(player.ID, 0, 10, bson.D{{"order", 1}})
	if err != nil {
		return message, err
	}

	if len(wishlistApps) > 10 {
		wishlistApps = wishlistApps[0:10]
	}

	if len(wishlistApps) > 0 {

		message.Content = "<@" + msg.Author.ID + ">"
		message.Embed = &discordgo.MessageEmbed{
			Title:  "Wishlist Items",
			URL:    "https://gamedb.online" + player.GetPath() + "#wishlist",
			Author: getAuthor(msg.Author.ID),
		}

		var code []string

		for k, app := range wishlistApps {

			avatar := app.GetIcon()
			if strings.HasPrefix(avatar, "/") {
				avatar = "https://gamedb.online" + avatar
			}

			if k == 0 {
				message.Embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: avatar}
			}

			space := ""
			if k < 9 && len(wishlistApps) > 9 {
				space = " "
			}

			code = append(code, strconv.Itoa(app.Order)+": "+space+app.GetName())
		}

		message.Embed.Description = "```" + strings.Join(code, "\n") + "```"

	} else {
		message.Content = player.GetName() + " has no wishlist items" // todo, dont do as content
	}

	return message, nil
}
