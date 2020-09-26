package chatbot

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/mongo"
)

type CommandGroupsTrending struct {
}

func (c CommandGroupsTrending) ID() string {
	return CGroupsTrending
}

func (CommandGroupsTrending) Regex() string {
	return `^[.|!]trending[\s-]?groups$`
}

func (CommandGroupsTrending) DisableCache() bool {
	return false
}

func (CommandGroupsTrending) PerProdCode() bool {
	return false
}

func (CommandGroupsTrending) Example() string {
	return ".trending groups"
}

func (CommandGroupsTrending) Description() template.HTML {
	return "Returns the top trending groups"
}

func (CommandGroupsTrending) Type() CommandType {
	return TypeGroup
}

func (CommandGroupsTrending) Output(msg *discordgo.MessageCreate, _ steamapi.ProductCC) (message discordgo.MessageSend, err error) {

	message.Embed = &discordgo.MessageEmbed{
		Title:  "Trending Groups",
		URL:    config.C.GameDBDomain + "/groups",
		Author: getAuthor(msg.Author.ID),
	}

	groups, err := mongo.TrendingGroups()
	if err != nil {
		return message, err
	}

	if len(groups) > 10 {
		groups = groups[0:10]
	}

	var code []string

	for k, group := range groups {

		if k == 0 {

			avatar := group.GetIcon()
			if strings.HasPrefix(avatar, "/") {
				avatar = "https://gamedb.online" + avatar
			}

			message.Embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: avatar}
		}

		code = append(code, fmt.Sprintf("%2d", k+1)+": "+group.GetName()+" ("+group.GetTrend()+")")
	}

	message.Embed.Description = "```" + strings.Join(code, "\n") + "```"

	return message, nil
}
