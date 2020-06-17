package chatbot

import (
	"html/template"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/mongo"
)

type CommandGroupsTrending struct {
}

func (c CommandGroupsTrending) ID() string {
	return CGroupsTrending
}

func (CommandGroupsTrending) Regex() string {
	return `^[.|!]trending[\s-]]groups$`
}

func (CommandGroupsTrending) DisableCache() bool {
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

func (CommandGroupsTrending) Output(msg *discordgo.MessageCreate) (message discordgo.MessageSend, err error) {

	message.Content = "<@" + msg.Author.ID + ">"
	message.Embed = &discordgo.MessageEmbed{
		Title:  "Trending Groups",
		URL:    "https://gamedb.online/groups",
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

		space := ""
		if k < 9 {
			space = " "
		}

		code = append(code, helpers.OrdinalComma(k+1)+". "+space+group.GetName()+" - "+humanize.Comma(group.Trending)+" trend value")
	}

	message.Embed.Description = "```" + strings.Join(code, "\n") + "```"

	return message, nil
}
