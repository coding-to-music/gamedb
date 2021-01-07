package chatbot

import (
	"errors"
	"html/template"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/chatbot/charts"
	"github.com/gamedb/gamedb/pkg/chatbot/interactions"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
)

type CommandGroup struct {
}

func (c CommandGroup) ID() string {
	return CGroup
}

func (CommandGroup) Regex() string {
	return `^[.|!](group|clan) (.*)`
}

func (CommandGroup) DisableCache() bool {
	return false
}

func (CommandGroup) PerProdCode() bool {
	return false
}

func (CommandGroup) Example() string {
	return ".group {group}"
}

func (CommandGroup) Description() template.HTML {
	return "Get info on a group"
}

func (CommandGroup) Type() CommandType {
	return TypeGroup
}

func (c CommandGroup) Slash() []interactions.InteractionOption {

	return []interactions.InteractionOption{
		{
			Name:        "group",
			Description: "The name or ID of the group",
			Type:        interactions.InteractionOptionTypeString,
			Required:    true,
		},
	}
}

func (c CommandGroup) Output(msg *discordgo.MessageCreate, _ steamapi.ProductCC) (message discordgo.MessageSend, err error) {

	matches := RegexCache[c.Regex()].FindStringSubmatch(msg.Message.Content)
	if len(matches) == 0 {
		return message, errors.New("invalid regex")
	}

	groups, _, _, err := elasticsearch.SearchGroups(0, 1, nil, matches[2], "")
	if err != nil {
		return message, err
	} else if len(groups) == 0 {
		message.Content = "Group **" + matches[2] + "** not found"
		return message, nil
	}

	group := groups[0]

	var abbr = group.GetAbbr()
	if abbr == "" {
		abbr = "-"
	}
	var headline = group.GetHeadline()
	if headline == "" {
		headline = "-"
	}

	message.Embed = &discordgo.MessageEmbed{
		Title:       group.GetName(),
		Description: headline,
		URL:         config.C.GameDBDomain + group.GetPath(),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: group.GetIcon(),
		},
		Footer: getFooter(),
		Color:  2664261,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Members",
				Value: humanize.Comma(int64(group.Members)),
			},
		},
		Image: &discordgo.MessageEmbedImage{
			URL: charts.GetGroupChart(c.ID(), group.ID),
		},
	}

	return message, nil
}
