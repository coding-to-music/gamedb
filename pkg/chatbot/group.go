package chatbot

import (
	"html/template"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/chatbot/charts"
	"github.com/gamedb/gamedb/pkg/elastic-search"
	"github.com/gamedb/gamedb/pkg/log"
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

func (CommandGroup) Example() string {
	return ".group {game}"
}

func (CommandGroup) Description() template.HTML {
	return "Get info on a group"
}

func (CommandGroup) Type() CommandType {
	return TypeGroup
}

func (c CommandGroup) Output(msg *discordgo.MessageCreate) (message discordgo.MessageSend, err error) {

	matches := RegexCache[c.Regex()].FindStringSubmatch(msg.Message.Content)

	groups, _, _, err := elastic_search.SearchGroups(0, 1, nil, matches[2], "")
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

	var image *discordgo.MessageEmbedImage
	url, width, height, err := charts.GetGroupChart(group.ID)
	if err != nil {
		log.Err(err)
	} else if url != "" {
		image = &discordgo.MessageEmbedImage{
			URL:    url,
			Width:  width,
			Height: height,
		}
	}

	message.Content = "<@" + msg.Author.ID + ">"
	message.Embed = &discordgo.MessageEmbed{
		Image: image,
		Title: group.GetName(),
		URL:   "https://gamedb.online" + group.GetPath(),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: group.GetIcon(),
		},
		Footer: getFooter(),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Headline",
				Value: headline,
			},
			{
				Name:  "Short Name",
				Value: abbr,
			},
			{
				Name:  "Members",
				Value: humanize.Comma(int64(group.Members)),
			},
		},
	}

	return message, nil
}
