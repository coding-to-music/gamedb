package chatbot

import (
	"regexp"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/mongo"
)

type CommandPlayer struct {
}

func (CommandPlayer) Regex() *regexp.Regexp {
	return regexp.MustCompile("^.(player|user) (.*)")
}

func (c CommandPlayer) Output(input string) (message discordgo.MessageSend, err error) {

	matches := c.Regex().FindStringSubmatch(input)

	player, err := mongo.SearchPlayer(matches[2], nil)
	if err != nil {
		return message, err
	}

	message.Embed = &discordgo.MessageEmbed{
		Title:  player.GetName(),
		URL:    "https://gamedb.online" + player.GetPath(),
		Author: author,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://gamedb.online" + player.GetAvatar(),
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Level",
				Value: humanize.Comma(int64(player.Level)),
			},
			{
				Name:  "Games",
				Value: humanize.Comma(int64(player.GamesCount)),
			},
			{
				Name:  "Playtime",
				Value: helpers.GetTimeLong(player.PlayTime, 3),
			},
			{
				Name:  "Friends",
				Value: humanize.Comma(int64(player.FriendsCount)),
			},
		},
	}

	return message, nil
}

func (CommandPlayer) Example() string {
	return ".player {player_name}"
}

func (CommandPlayer) Description() string {
	return "Get info on a player"
}

func (CommandPlayer) Type() CommandType {
	return TypePlayer
}
