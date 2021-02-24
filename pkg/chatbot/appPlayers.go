package chatbot

import (
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/chatbot/charts"
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

func (CommandAppPlayers) Description() string {
	return "Retrieve information about the number of people playing a game"
}

func (CommandAppPlayers) Type() CommandType {
	return TypeGame
}

func (c CommandAppPlayers) LegacyInputs(input string) map[string]string {

	matches := RegexCache[c.Regex()].FindStringSubmatch(input)

	return map[string]string{
		"game": matches[2],
	}
}

func (c CommandAppPlayers) Slash() []*discordgo.ApplicationCommandOption {

	return []*discordgo.ApplicationCommandOption{
		{
			Name:        "game",
			Description: "The name or ID of the game, or blank for all of steam",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    true,
		},
	}
}

func (c CommandAppPlayers) Output(_ string, _ steamapi.ProductCC, inputs map[string]string) (message discordgo.MessageSend, err error) {

	if inputs["game"] == "" {
		message.Content = "Missing game name"
		return message, nil
	}

	apps, err := elasticsearch.SearchAppsSimple(1, inputs["game"])
	if err != nil {
		return message, err
	} else if len(apps) == 0 {
		message.Content = "Game **" + inputs["game"] + "** not found on Steam"
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
		URL:       app.GetPathAbsolute(),
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: app.GetHeaderImage(), Width: 460, Height: 215},
		Footer:    getFooter(),
		Color:     greenHexDec,
		Image:     &discordgo.MessageEmbedImage{URL: charts.GetAppPlayersChart(c.ID(), app.ID, "10m", "7d", "Players (1 Week)")},
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
