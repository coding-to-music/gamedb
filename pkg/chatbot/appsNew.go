package chatbot

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/chatbot/interactions"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/mongo"
)

type CommandAppsNew struct {
}

func (c CommandAppsNew) ID() string {
	return CAppsNew
}

func (CommandAppsNew) Regex() string {
	return `^[.|!]new$`
}

func (CommandAppsNew) DisableCache() bool {
	return false
}

func (CommandAppsNew) PerProdCode() bool {
	return false
}

func (CommandAppsNew) Example() string {
	return ".new"
}

func (CommandAppsNew) Description() template.HTML {
	return "Returns the most popular newly released games"
}

func (CommandAppsNew) Type() CommandType {
	return TypeGame
}

func (CommandAppsNew) LegacyPrefix() string {
	return "new$"
}

func (c CommandAppsNew) Slash() []interactions.InteractionOption {
	return []interactions.InteractionOption{}
}

func (CommandAppsNew) Output(msg *discordgo.MessageCreate, _ steamapi.ProductCC) (message discordgo.MessageSend, err error) {

	message.Embed = &discordgo.MessageEmbed{
		Title:  "Popular New Apps",
		URL:    config.C.GameDBDomain + "/games/new-releases",
		Author: getAuthor(msg.Author.ID),
		Color:  2664261,
	}

	apps, err := mongo.PopularNewApps()
	if err != nil {
		return message, err
	}

	if len(apps) > 10 {
		apps = apps[0:10]
	}

	var code []string

	for k, v := range apps {

		if k == 0 {
			message.Embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: v.GetHeaderImage()}
		}

		code = append(code, fmt.Sprintf("%2d", k+1)+": "+humanize.Comma(int64(v.PlayerPeakWeek))+" - "+v.GetName())
	}

	message.Embed.Description = "```" + strings.Join(code, "\n") + "```"

	return message, nil
}
