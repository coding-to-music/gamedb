package chatbot

import (
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/chatbot/charts"
	"github.com/gamedb/gamedb/pkg/chatbot/interactions"
	"github.com/gamedb/gamedb/pkg/mongo"
)

type CommandSteamOnline struct {
}

func (c CommandSteamOnline) ID() string {
	return CSteamOnline
}

func (CommandSteamOnline) Regex() string {
	return `^[.|!](players|online)$`
}

func (CommandSteamOnline) DisableCache() bool {
	return false
}

func (CommandSteamOnline) PerProdCode() bool {
	return false
}

func (CommandSteamOnline) Example() string {
	return ".players"
}

func (CommandSteamOnline) Description() string {
	return "Retrieve the number of people currently on Steam"
}

func (CommandSteamOnline) Type() CommandType {
	return TypeOther
}

func (CommandSteamOnline) LegacyInputs(input string) map[string]string {
	return map[string]string{}
}

func (c CommandSteamOnline) Slash() []interactions.InteractionOption {

	return []interactions.InteractionOption{}
}

func (c CommandSteamOnline) Output(_ string, _ steamapi.ProductCC, _ map[string]string) (message discordgo.MessageSend, err error) {

	var app = mongo.App{}

	i, err := app.GetPlayersOnline()
	if err != nil {
		return message, err
	}

	i2, err := app.GetPlayersInGame()
	if err != nil {
		return message, err
	}

	message.Embed = &discordgo.MessageEmbed{
		Title:     "Online Players",
		URL:       "https://gamedb.online/stats",
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: "https://gamedb.online/assets/img/no-app-image-square.jpg"},
		Footer:    getFooter(),
		Color:     greenHexDec,
		Image:     &discordgo.MessageEmbedImage{URL: charts.GetAppPlayersChart(c.ID(), 0, "7d", "10m")},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Online",
				Value:  humanize.Comma(i),
				Inline: true,
			},
			{
				Name:   "In Game",
				Value:  humanize.Comma(i2),
				Inline: true,
			},
		},
	}

	return message, nil
}
