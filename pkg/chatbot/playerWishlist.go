package chatbot

import (
	"fmt"
	"strconv"
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

type CommandPlayerWishlist struct {
}

func (c CommandPlayerWishlist) ID() string {
	return CPlayerWishlist
}

func (CommandPlayerWishlist) Regex() string {
	return `^[.|!]wishlist (.{2,32})`
}

func (CommandPlayerWishlist) DisableCache() bool {
	return false
}

func (CommandPlayerWishlist) PerProdCode() bool {
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

func (c CommandPlayerWishlist) LegacyInputs(input string) map[string]string {

	matches := RegexCache[c.Regex()].FindStringSubmatch(input)

	return map[string]string{
		"player": matches[1],
	}
}

func (c CommandPlayerWishlist) Slash() []interactions.InteractionOption {

	return []interactions.InteractionOption{
		{
			Name:        "player",
			Description: "The name or ID of the player",
			Type:        interactions.InteractionOptionTypeString,
			Required:    true,
		},
	}
}

func (c CommandPlayerWishlist) Output(authorID string, _ steamapi.ProductCC, inputs map[string]string) (message discordgo.MessageSend, err error) {

	player, q, err := mongo.SearchPlayer(inputs["player"], nil)
	if err == mongo.ErrNoDocuments {

		message.Content = "Player **" + inputs["player"] + "** not found, please enter a user's vanity URL"
		return message, nil

	} else if err != nil {
		return message, err
	}

	if q {
		err = queue.ProducePlayer(queue.PlayerMessage{ID: player.ID}, "chatbot-player.wishlist")
		err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
		if err != nil {
			log.ErrS(err)
		}
	}

	wishlistApps, err := mongo.GetPlayerWishlistAppsByPlayer(player.ID, 0, 10, bson.D{{"order", 1}}, nil)
	if err != nil {
		return message, err
	}

	if len(wishlistApps) > 10 {
		wishlistApps = wishlistApps[0:10]
	}

	if len(wishlistApps) > 0 {

		message.Embed = &discordgo.MessageEmbed{
			Title:  "Wishlist Items",
			URL:    config.C.GameDBDomain + player.GetPath() + "#wishlist",
			Author: getAuthor(authorID),
			Color:  2664261,
		}

		var code []string

		for k, app := range wishlistApps {

			if k == 0 {

				avatar := app.GetHeaderImage()
				if strings.HasPrefix(avatar, "/") {
					avatar = "https://gamedb.online" + avatar
				}

				message.Embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: avatar}
			}

			var rank string
			if app.Order > 0 {
				rank = strconv.Itoa(app.Order)
			} else {
				rank = "*"
			}

			code = append(code, fmt.Sprintf("%2s", rank)+": "+app.GetName())
		}

		message.Embed.Description = "```" + strings.Join(code, "\n") + "```"

	} else {
		message.Content = player.GetName() + " has no wishlist items, or a profile set to private"
	}

	return message, nil
}
