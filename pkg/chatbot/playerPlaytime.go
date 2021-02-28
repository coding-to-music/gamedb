package chatbot

import (
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/chatbot/charts"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/helpers"
)

type CommandPlayerPlaytime struct {
}

func (c CommandPlayerPlaytime) ID() string {
	return CPlayerPlaytime
}

func (CommandPlayerPlaytime) Regex() string {
	return `^[.|!]playtime (.+)`
}

func (CommandPlayerPlaytime) DisableCache() bool {
	return false
}

func (CommandPlayerPlaytime) PerProdCode() bool {
	return false
}

func (CommandPlayerPlaytime) AllowDM() bool {
	return false
}

func (CommandPlayerPlaytime) Example() string {
	return ".playtime {player}"
}

func (CommandPlayerPlaytime) Description() string {
	return "Retrieve a player's total playtime"
}

func (CommandPlayerPlaytime) Type() CommandType {
	return TypePlayer
}

func (c CommandPlayerPlaytime) LegacyInputs(input string) map[string]string {

	matches := RegexCache[c.Regex()].FindStringSubmatch(input)

	return map[string]string{
		"player": matches[1],
	}
}

func (c CommandPlayerPlaytime) Slash() []*discordgo.ApplicationCommandOption {

	return []*discordgo.ApplicationCommandOption{
		{
			Name:        "player",
			Description: "The name or ID of the player",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    true,
		},
	}
}

func (c CommandPlayerPlaytime) Output(_ string, _ steamapi.ProductCC, inputs map[string]string) (message discordgo.MessageSend, err error) {

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
	if val, ok := player.Ranks[helpers.RankKeyPlaytime]; ok && val > 0 {
		rank = helpers.OrdinalComma(val)
	}

	if player.PlayTime == 0 {
		message.Content = "Profile set to private"
	} else {

		message.Embed = &discordgo.MessageEmbed{
			Title:     player.GetName(),
			URL:       player.GetPathAbsolute(),
			Thumbnail: &discordgo.MessageEmbedThumbnail{URL: player.GetAvatarAbsolute(), Width: 184, Height: 184},
			Footer:    getFooter(),
			Color:     greenHexDec,
			Image:     &discordgo.MessageEmbedImage{URL: charts.GetPlayerChart(c.ID(), player.ID, helpers.InfPlayersPlaytime, "Playtime")},
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Playtime",
					Value:  helpers.GetTimeLong(player.PlayTime, 0),
					Inline: true,
				},
				{
					Name:   "Rank",
					Value:  rank,
					Inline: true,
				},
			},
		}
	}

	return message, nil
}
