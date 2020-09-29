package chatbot

import (
	"html/template"
	"strings"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"go.mongodb.org/mongo-driver/bson"
)

type CommandPlayer struct {
}

func (c CommandPlayer) ID() string {
	return CPlayer
}

func (CommandPlayer) Regex() string {
	return `^[.|!](player|user) (.{2,32})$`
}

func (CommandPlayer) DisableCache() bool {
	return false
}

func (CommandPlayer) PerProdCode() bool {
	return false
}

func (CommandPlayer) Example() string {
	return ".player {player}"
}

func (CommandPlayer) Description() template.HTML {
	return "Get info on a player"
}

func (CommandPlayer) Type() CommandType {
	return TypePlayer
}

func (c CommandPlayer) Output(msg *discordgo.MessageCreate, _ steamapi.ProductCC) (message discordgo.MessageSend, err error) {

	matches := RegexCache[c.Regex()].FindStringSubmatch(msg.Message.Content)

	projection := bson.M{
		"_id":               1,
		"achievement_count": 1,
		"avatar":            1,
		"badges_count":      1,
		"badges_foil_count": 1,
		"comments_count":    1,
		"friends_count":     1,
		"games_count":       1,
		"level":             1,
		"persona_name":      1,
		"play_time":         1,
		"ranks":             1,
	}

	player, q, err := mongo.SearchPlayer(matches[2], projection)
	if err == mongo.ErrNoDocuments {

		message.Content = "Player **" + matches[2] + "** not found, please enter a user's vanity URL"
		return message, nil

	} else if err != nil {
		return message, err
	}

	if q {
		err = queue.ProducePlayer(queue.PlayerMessage{ID: player.ID})
		err = helpers.IgnoreErrors(err, memcache.ErrInQueue)
		if err != nil {
			log.ErrS(err)
		}
	}

	avatar := player.GetAvatar()
	if strings.HasPrefix(avatar, "/") {
		avatar = "https://gamedb.online" + avatar
	}

	var games string
	if player.GamesCount == 0 {
		games = "<private profile>"
	} else {
		games = humanize.Comma(int64(player.GamesCount))
	}

	var playtime string
	if player.PlayTime == 0 {
		playtime = "<private profile>"
	} else {
		playtime = helpers.GetTimeLong(player.PlayTime, 3)
	}

	message.Embed = &discordgo.MessageEmbed{
		Title: player.GetName(),
		URL:   config.C.GameDBDomain + player.GetPath(),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: avatar,
		},
		Footer: getFooter(),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Level",
				Value: humanize.Comma(int64(player.Level)) + " (" + helpers.OrdinalComma(player.Ranks[string(mongo.RankKeyLevel)]) + ")",
			},
			{
				Name:  "Games",
				Value: games + " (" + helpers.OrdinalComma(player.Ranks[string(mongo.RankKeyGames)]) + ")",
			},
			{
				Name:  "Achievements",
				Value: humanize.Comma(int64(player.AchievementCount)) + " (" + helpers.OrdinalComma(player.Ranks[string(mongo.RankKeyAchievements)]) + ")",
			},
			{
				Name:  "Badges",
				Value: humanize.Comma(int64(player.BadgesCount)) + " (" + helpers.OrdinalComma(player.Ranks[string(mongo.RankKeyBadges)]) + ")",
			},
			{
				Name:  "Foil Badges",
				Value: humanize.Comma(int64(player.BadgesFoilCount)) + " (" + helpers.OrdinalComma(player.Ranks[string(mongo.RankKeyBadgesFoil)]) + ")",
			},
			{
				Name:  "Playtime",
				Value: playtime + " (" + helpers.OrdinalComma(player.Ranks[string(mongo.RankKeyPlaytime)]) + ")",
			},
			{
				Name:  "Friends",
				Value: humanize.Comma(int64(player.FriendsCount)) + " (" + helpers.OrdinalComma(player.Ranks[string(mongo.RankKeyFriends)]) + ")",
			},
			{
				Name:  "Comments",
				Value: humanize.Comma(int64(player.CommentsCount)) + " (" + helpers.OrdinalComma(player.Ranks[string(mongo.RankKeyComments)]) + ")",
			},
		},
	}

	return message, nil
}
