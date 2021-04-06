package chatbot

import (
	"strconv"
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/oauth"
)

type CommandPlayer struct {
}

func (c CommandPlayer) ID() string {
	return CPlayer
}

func (CommandPlayer) Regex() string {
	return `^[.|!](player|user)\s(.+)`
}

func (CommandPlayer) DisableCache() bool {
	return false
}

func (CommandPlayer) PerProdCode() bool {
	return false
}

func (CommandPlayer) AllowDM() bool {
	return false
}

func (CommandPlayer) Example() string {
	return ".player {player}?"
}

func (CommandPlayer) Description() string {
	return "Retrieve information about a player"
}

func (CommandPlayer) Type() CommandType {
	return TypePlayer
}

func (c CommandPlayer) LegacyInputs(input string) map[string]string {

	matches := RegexCache[c.Regex()].FindStringSubmatch(input)

	return map[string]string{
		"player": matches[2],
	}
}

func (c CommandPlayer) Slash() []*discordgo.ApplicationCommandOption {

	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "player",
			Description: "The name or ID of the player",
			Required:    false,
		},
	}
}

func (c CommandPlayer) Output(authorID string, _ steamapi.ProductCC, inputs map[string]string) (message discordgo.MessageSend, err error) {

	var player Player

	if inputs["player"] == "" {

		provider, err := mysql.GetUserProviderByProviderID(oauth.ProviderDiscord, authorID)
		if err != nil {
			message.Content = "Please connect your Discord account first: <" + config.C.GlobalSteamDomain + "/oauth/out/discord?page=settings>"
			return message, nil
		}

		provider, err = mysql.GetUserProviderByUserID(oauth.ProviderSteam, provider.UserID)
		if err != nil {
			message.Content = "Please connect your Steam account first: <" + config.C.GlobalSteamDomain + "/oauth/out/steam?page=settings>"
			return message, nil
		}

		i, err := strconv.ParseInt(provider.ID, 10, 64)
		if err != nil || i == 0 {
			message.Content = "We had trouble finding your profile on Global Steam"
			return message, nil
		}

		player, err = mongo.GetPlayer(i)
		if err != nil {
			message.Content = "We had trouble finding your profile on Global Steam"
			return message, nil
		}

	} else {

		player, err = searchForPlayer(inputs["player"])
		if err == elasticsearch.ErrNoResult || err == steamapi.ErrProfileMissing {

			message.Content = "Player **" + inputs["player"] + "** not found, they may be set to private, please enter a user's vanity URL"
			return message, nil

		} else if err != nil {
			return message, err
		}
	}

	// Fields
	var games = "None / Profile set to private"
	if player.GetGamesCount() > 0 {
		games = humanize.Comma(int64(player.GetGamesCount()))
		if val, ok := player.GetRanks()[helpers.RankKeyGames]; ok && val > 0 {
			games += " (" + helpers.OrdinalComma(val) + ")"
		} else {
			games += " (Unranked)"
		}
	}

	var level = "Profile set to private"
	if player.GetLevel() > 0 {
		level = humanize.Comma(int64(player.GetLevel()))
		if val, ok := player.GetRanks()[helpers.RankKeyLevel]; ok && val > 0 {
			level += " (" + helpers.OrdinalComma(val) + ")"
		} else {
			level += " (Unranked)"
		}
	}

	var badges = "None / Profile set to private"
	if player.GetBadges() > 0 {
		badges = humanize.Comma(int64(player.GetBadges()))
		if val, ok := player.GetRanks()[helpers.RankKeyBadges]; ok && val > 0 {
			badges += " (" + helpers.OrdinalComma(val) + ")"
		} else {
			badges += " (Unranked)"
		}
	}

	var foils = "None / Profile set to private"
	if player.GetBadgesFoil() > 0 {
		foils = humanize.Comma(int64(player.GetBadgesFoil()))
		if val, ok := player.GetRanks()[helpers.RankKeyBadgesFoil]; ok && val > 0 {
			foils += " (" + helpers.OrdinalComma(val) + ")"
		} else {
			foils += " (Unranked)"
		}
	}

	var achievements = "None / Profile set to private"
	if player.GetAchievements() > 0 {
		achievements = humanize.Comma(int64(player.GetAchievements()))
		if val, ok := player.GetRanks()[helpers.RankKeyAchievements]; ok && val > 0 {
			achievements += " (" + helpers.OrdinalComma(val) + ")"
		} else {
			achievements += " (Unranked)"
		}
	}

	var playtime = "None / Profile set to private"
	if player.GetPlaytime() > 0 {
		playtime = helpers.GetTimeLong(player.GetPlaytime(), 3)
		if val, ok := player.GetRanks()[helpers.RankKeyPlaytime]; ok && val > 0 {
			playtime += " (" + helpers.OrdinalComma(val) + ")"
		} else {
			playtime += " (Unranked)"
		}
	}

	var bans = humanize.Comma(int64(player.GetGameBans())) + " Game Bans\n" + humanize.Comma(int64(player.GetVACBans())) + " VAC Bans"
	if player.GetVACBans() > 0 {
		bans += "\nVAC Banned " + helpers.GetTimeShort(int(time.Since(player.GetLastBan()).Minutes()), 2) + " ago"
	}

	//
	message.Embed = &discordgo.MessageEmbed{
		Title:     player.GetName(),
		URL:       player.GetPathAbsolute(),
		Thumbnail: &discordgo.MessageEmbedThumbnail{URL: player.GetAvatarAbsolute(), Width: 184, Height: 184},
		Footer:    getFooter(),
		Color:     greenHexDec,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Level",
				Value: level,
			},
			{
				Name:  "Games",
				Value: games,
			},
			{
				Name:  "Achievements",
				Value: achievements,
			},
			{
				Name:  "Badges",
				Value: badges,
			},
			{
				Name:  "Foil Badges",
				Value: foils,
			},
			{
				Name:  "Playtime",
				Value: playtime,
			},
			{
				Name:  "Bans",
				Value: bans,
			},
		},
	}

	return message, nil
}
