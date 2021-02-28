package chatbot

import (
	"fmt"
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
	return `^[.|!]trending[\s-]?groups`
}

func (CommandGroupsTrending) DisableCache() bool {
	return false
}

func (CommandGroupsTrending) PerProdCode() bool {
	return false
}

func (CommandGroupsTrending) AllowDM() bool {
	return false
}

func (CommandGroupsTrending) Example() string {
	return ".trending groups"
}

func (CommandGroupsTrending) Description() string {
	return "Retrieve the most trending groups"
}

func (CommandGroupsTrending) Type() CommandType {
	return TypeGroup
}

func (CommandGroupsTrending) LegacyInputs(_ string) map[string]string {
	return map[string]string{}
}

func (c CommandGroupsTrending) Slash() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{}
}

func (CommandGroupsTrending) Output(authorID string, _ steamapi.ProductCC, _ map[string]string) (message discordgo.MessageSend, err error) {

	message.Embed = &discordgo.MessageEmbed{
		Title:  "Trending Groups",
		URL:    config.C.GameDBDomain + "/groups",
		Author: getAuthor(authorID),
		Color:  greenHexDec,
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
			message.Embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: group.GetIconAbsolute()}
		}

		code = append(code, fmt.Sprintf("%2d", k+1)+": "+group.GetTrend()+" - "+group.GetName())
	}

	message.Embed.Description = "```" + strings.Join(code, "\n") + "```"

	return message, nil
}
