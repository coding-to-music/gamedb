package chatbot

import (
	"strconv"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/chatbot/charts"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/influx/schemas"
)

type CommandPlayerApps struct {
}

func (c CommandPlayerApps) ID() string {
	return CPlayerApps
}

func (CommandPlayerApps) Regex() string {
	return `^[.|!](games|apps) (.+)`
}

func (CommandPlayerApps) DisableCache() bool {
	return false
}

func (CommandPlayerApps) PerProdCode() bool {
	return false
}

func (CommandPlayerApps) AllowDM() bool {
	return false
}

func (CommandPlayerApps) Example() string {
	return ".games {player}"
}

func (CommandPlayerApps) Description() string {
	return "Retrieve the number of games in a player's library"
}

func (CommandPlayerApps) Type() CommandType {
	return TypePlayer
}

func (c CommandPlayerApps) LegacyInputs(input string) map[string]string {

	matches := RegexCache[c.Regex()].FindStringSubmatch(input)

	return map[string]string{
		"player": matches[2],
	}
}

func (c CommandPlayerApps) Slash() []*discordgo.ApplicationCommandOption {

	return []*discordgo.ApplicationCommandOption{
		{
			Name:        "player",
			Description: "The name or ID of the player",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    true,
		},
	}
}

func (c CommandPlayerApps) Output(_ string, _ steamapi.ProductCC, inputs map[string]string) (message discordgo.MessageSend, err error) {

	if inputs["player"] == "" {
		message.Content = "Missing player name"
		return message, nil
	}

	player, err := searchForPlayer(inputs["player"])
	if err == elasticsearch.ErrNoResult || err == steamapi.ErrProfileMissing {

		message.Content = "Player **" + inputs["player"] + "** not found, they may be set to private, please enter a user's vanity URL"
		return message, nil

	} else if err != nil {
		return message, err
	}

	// Sucess response
	var rank = "Unranked"
	if val, ok := player.Ranks[helpers.RankKeyGames]; ok && val > 0 {
		rank = helpers.OrdinalComma(val)
	}

	if player.Games > 0 {
		message.Embed = &discordgo.MessageEmbed{
			Title:     player.GetName(),
			URL:       player.GetPathAbsolute(),
			Thumbnail: &discordgo.MessageEmbedThumbnail{URL: player.GetAvatarAbsolute(), Width: 184, Height: 184},
			Footer:    getFooter(),
			Color:     greenHexDec,
			Image:     &discordgo.MessageEmbedImage{URL: charts.GetPlayerChart(c.ID(), player.ID, schemas.InfPlayersGames, "Games")},
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Games",
					Value:  strconv.Itoa(player.Games),
					Inline: true,
				},
				{
					Name:   "Rank",
					Value:  rank,
					Inline: true,
				},
			},
		}
	} else {
		message.Content = "Profile set to private"
	}

	return message, nil
}
