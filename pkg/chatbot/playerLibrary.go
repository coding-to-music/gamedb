package chatbot

import (
	"fmt"
	"strings"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type CommandPlayerLibrary struct {
}

func (c CommandPlayerLibrary) ID() string {
	return CPlayerLibrary
}

func (CommandPlayerLibrary) Regex() string {
	return `^[.|!](library|top) (.+)`
}

func (CommandPlayerLibrary) DisableCache() bool {
	return false
}

func (CommandPlayerLibrary) PerProdCode() bool {
	return false
}

func (CommandPlayerLibrary) AllowDM() bool {
	return false
}

func (CommandPlayerLibrary) Example() string {
	return ".library {player}"
}

func (CommandPlayerLibrary) Description() string {
	return "Retrieve a players top played games"
}

func (CommandPlayerLibrary) Type() CommandType {
	return TypePlayer
}

func (c CommandPlayerLibrary) LegacyInputs(input string) map[string]string {

	matches := RegexCache[c.Regex()].FindStringSubmatch(input)

	return map[string]string{
		"player": matches[2],
	}
}

func (c CommandPlayerLibrary) Slash() []*discordgo.ApplicationCommandOption {

	return []*discordgo.ApplicationCommandOption{
		{
			Name:        "player",
			Description: "The name or ID of the player",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    true,
		},
	}
}

func (c CommandPlayerLibrary) Output(authorID string, _ steamapi.ProductCC, inputs map[string]string) (message discordgo.MessageSend, err error) {

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
	apps, err := mongo.GetPlayerAppsByPlayer(player.ID, 0, 10, bson.D{{"app_time", -1}}, bson.M{"app_name": 1, "app_time": 1, "app_id": 1}, nil)
	if err != nil {
		return message, err
	}

	if len(apps) > 10 {
		apps = apps[0:10]
	}

	if len(apps) > 0 {

		var code []string
		for k, app := range apps {
			code = append(code, fmt.Sprintf("%2d", k+1)+": "+app.GetTimeNice()+" - "+app.AppName)
		}

		message.Embed = &discordgo.MessageEmbed{
			Title:       player.GetName() + "'s Top Games",
			URL:         player.GetPathAbsolute() + "#games",
			Author:      getAuthor(authorID),
			Color:       greenHexDec,
			Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: player.GetAvatarAbsolute(), Width: 184, Height: 184},
			Description: "```" + strings.Join(code, "\n") + "```",
		}

	} else {
		message.Content = "Profile set to private"
	}

	return message, nil
}
