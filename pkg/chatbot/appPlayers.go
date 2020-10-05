package chatbot

import (
	"html/template"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/chatbot/charts"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/mongo"
)

type CommandAppPlayers struct {
}

func (c CommandAppPlayers) ID() string {
	return CAppPlayers
}

func (CommandAppPlayers) Regex() string {
	return `^[.|!](players|online) (.*)`
}

func (CommandAppPlayers) DisableCache() bool {
	return false
}

func (CommandAppPlayers) PerProdCode() bool {
	return false
}

func (CommandAppPlayers) Example() string {
	return ".players {game}"
}

func (CommandAppPlayers) Description() template.HTML {
	return "Gets the number of people playing a game."
}

func (CommandAppPlayers) Type() CommandType {
	return TypeGame
}

func (c CommandAppPlayers) Output(msg *discordgo.MessageCreate, _ steamapi.ProductCC) (message discordgo.MessageSend, err error) {

	matches := RegexCache[c.Regex()].FindStringSubmatch(msg.Message.Content)

	apps, err := elasticsearch.SearchAppsSimple(1, matches[2])
	if err != nil {
		return message, err
	} else if len(apps) == 0 {
		message.Content = "Game **" + matches[2] + "** not found"
		return message, nil
	}

	app, err := mongo.GetApp(apps[0].ID)
	if err != nil {
		return message, err
	}

	i, err := app.GetPlayersInGame()
	if err != nil {
		return message, err
	}

	message.Embed = &discordgo.MessageEmbed{
		Title:     app.GetName(),
		URL:       config.C.GameDBDomain + app.GetPath(),
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: app.GetHeaderImage()},
		Footer:    getFooter(),
		Image: &discordgo.MessageEmbedImage{
			URL: charts.GetAppPlayersChart(c.ID(), app.ID, "7d", "30m"),
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Now",
				Value:  humanize.Comma(i),
				Inline: true,
			},
			{
				Name:   "7 Days",
				Value:  humanize.Comma(int64(app.PlayerPeakWeek)),
				Inline: true,
			},
			{
				Name:   "All Time",
				Value:  humanize.Comma(int64(app.PlayerPeakAllTime)),
				Inline: true,
			},
		},
	}

	return message, nil
}
