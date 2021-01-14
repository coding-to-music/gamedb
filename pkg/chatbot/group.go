package chatbot

import (
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

func (CommandGroup) Description() string {
	return "Retrieve information about a group"
}

func (CommandGroup) Type() CommandType {
	return TypeGroup
}

func (c CommandGroup) LegacyInputs(input string) map[string]string {

	matches := RegexCache[c.Regex()].FindStringSubmatch(input)

	return map[string]string{
		"group": matches[2],
	}
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

func (c CommandGroup) Output(_ string, _ steamapi.ProductCC, inputs map[string]string) (message discordgo.MessageSend, err error) {

	groups, _, _, err := elasticsearch.SearchGroups(0, 1, nil, inputs["group"], "")
	if err != nil {
		return message, err
	} else if len(groups) == 0 {
		message.Content = "Group **" + inputs["group"] + "** not found"
		return message, nil
	}

	var abbr = groups[0].GetAbbr()
	if abbr == "" {
		abbr = "-"
	}
	var headline = groups[0].GetHeadline()
	if headline == "" {
		headline = "-"
	}

	message.Embed = &discordgo.MessageEmbed{
		Title:       groups[0].GetName(),
		Description: headline,
		URL:         config.C.GameDBDomain + groups[0].GetPath(),
		Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: groups[0].GetIcon()},
		Footer:      getFooter(),
		Color:       greenHexDec,
		Image:       &discordgo.MessageEmbedImage{URL: charts.GetGroupChart(c.ID(), groups[0].ID, "Members")},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Members",
				Value: humanize.Comma(int64(groups[0].Members)),
			},
		},
	}

	return message, nil
}
