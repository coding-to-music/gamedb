package chatbot

import (
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/chatbot/charts"
	"github.com/gamedb/gamedb/pkg/chatbot/interactions"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/mongo"
)

type CommandPlayerPlaytime struct {
}

func (c CommandPlayerPlaytime) ID() string {
	return CPlayerPlaytime
}

func (CommandPlayerPlaytime) Regex() string {
	return `^[.|!]playtime (.{2,32})`
}

func (CommandPlayerPlaytime) DisableCache() bool {
	return false
}

func (CommandPlayerPlaytime) PerProdCode() bool {
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

func (c CommandPlayerPlaytime) Slash() []interactions.InteractionOption {

	return []interactions.InteractionOption{
		{
			Name:        "player",
			Description: "The name or ID of the player",
			Type:        interactions.InteractionOptionTypeString,
			Required:    true,
		},
	}
}

func (c CommandPlayerPlaytime) Output(_ string, _ steamapi.ProductCC, inputs map[string]string) (message discordgo.MessageSend, err error) {

	player, err := searchForPlayer(inputs["player"])
	if err == mongo.ErrNoDocuments {

		message.Content = "Player **" + inputs["player"] + "** not found, please enter a user's vanity URL"
		return message, nil

	} else if err != nil {
		return message, err
	}

	// Sucess response
	var rank = "Unranked"
	if val, ok := player.Ranks[helpers.RankKeyPlaytime]; ok {
		rank = helpers.OrdinalComma(val)
	}

	if player.PlayTime == 0 {
		message.Content = "Profile set to private"
	} else {

		message.Embed = &discordgo.MessageEmbed{
			Title:     player.GetName(),
			URL:       config.C.GameDBDomain + player.GetPath(),
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
