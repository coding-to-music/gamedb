package chatbot

import (
	"regexp"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/mongo"
)

type CommandGroup struct {
}

func (CommandGroup) Regex() *regexp.Regexp {
	return regexp.MustCompile(`^[.|!](group|clan) (.*)`)
}

func (c CommandGroup) Output(msg *discordgo.MessageCreate) (message discordgo.MessageSend, err error) {

	matches := c.Regex().FindStringSubmatch(msg.Message.Content)

	group, err := mongo.SearchGroups(matches[2])
	if err == mongo.ErrNoDocuments {

		message.Content = "Group **" + matches[2] + "** not found"
		return message, nil

	} else if err != nil {
		return message, err
	}

	if group.Abbr == "" {
		group.Abbr = "-"
	}

	if group.Headline == "" {
		group.Headline = "-"
	}

	message.Content = "<@" + msg.Author.ID + ">"
	message.Embed = &discordgo.MessageEmbed{
		Title: group.GetName(),
		URL:   "https://gamedb.online" + group.GetPath(),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: group.GetIcon(),
		},
		Footer: getFooter(),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Headline",
				Value: group.Headline,
			},
			{
				Name:  "Short Name",
				Value: group.Abbr,
			},
			{
				Name:  "Members",
				Value: humanize.Comma(int64(group.Members)),
			},
			{
				Name:  "Trend",
				Value: helpers.TrendValue(group.Trending),
			},
		},
	}

	return message, nil
}

func (CommandGroup) Example() string {
	return ".group {group_name}"
}

func (CommandGroup) Description() string {
	return "Get info on a group"
}

func (CommandGroup) Type() CommandType {
	return TypeGroup
}
