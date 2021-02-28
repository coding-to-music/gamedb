package chatbot

import (
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/chatbot/charts"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/helpers"
)

type CommandPlayerLevel struct {
}

func (c CommandPlayerLevel) ID() string {
	return CPlayerLevel
}

func (CommandPlayerLevel) Regex() string {
	return `^[.|!]level (.+)`
}

func (CommandPlayerLevel) DisableCache() bool {
	return false
}

func (CommandPlayerLevel) PerProdCode() bool {
	return false
}

func (CommandPlayerLevel) AllowDM() bool {
	return false
}

func (CommandPlayerLevel) Example() string {
	return ".level {player}"
}

func (CommandPlayerLevel) Description() string {
	return "Retrieve a player's level"
}

func (CommandPlayerLevel) Type() CommandType {
	return TypePlayer
}

func (c CommandPlayerLevel) LegacyInputs(input string) map[string]string {

	matches := RegexCache[c.Regex()].FindStringSubmatch(input)

	return map[string]string{
		"player": matches[1],
	}
}

func (c CommandPlayerLevel) Slash() []*discordgo.ApplicationCommandOption {

	return []*discordgo.ApplicationCommandOption{
		{
			Name:        "player",
			Description: "The name or ID of the player",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    true,
		},
	}
}

func (c CommandPlayerLevel) Output(_ string, _ steamapi.ProductCC, inputs map[string]string) (message discordgo.MessageSend, err error) {

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
	if val, ok := player.Ranks[helpers.RankKeyLevel]; ok && val > 0 {
		rank = helpers.OrdinalComma(val)
	}

	message.Embed = &discordgo.MessageEmbed{
		Title:     player.GetName(),
		URL:       player.GetPathAbsolute(),
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: player.GetAvatarAbsolute(), Width: 184, Height: 184},
		Footer:    getFooter(),
		Color:     greenHexDec,
		Image:     &discordgo.MessageEmbedImage{URL: charts.GetPlayerChart(c.ID(), player.ID, helpers.InfPlayersLevel, "Level")},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Level",
				Value:  humanize.Comma(int64(player.Level)),
				Inline: true,
			},
			{
				Name:   "Rank",
				Value:  rank,
				Inline: true,
			},
		},
	}

	return message, nil
}
