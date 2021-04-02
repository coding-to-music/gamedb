package helpers

import (
	"math"
	"strconv"
	"strings"

	"github.com/Jleagle/steam-go/steamid"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gosimple/slug"
)

const (
	DefaultPlayerAvatar = "/assets/img/no-player-image.jpg"
)

// Just use ID in case slug has changed
func GetPlayerCommunityLink(playerID int64) string {
	return "https://steamcommunity.com/profiles/" + strconv.FormatInt(playerID, 10)
}

// 184 x 184
func GetPlayerAvatar(avatar string) string {

	if strings.HasPrefix(avatar, "http") || strings.HasPrefix(avatar, "/") {
		return avatar
	} else if RegexSha1Only.MatchString(avatar) {
		return AvatarBase + avatar[0:2] + "/" + avatar + "_full.jpg"
	} else {
		return DefaultPlayerAvatar
	}
}

func GetPlayerAvatarAbsolute(avatar string) string {
	avatar = GetPlayerAvatar(avatar)
	if strings.HasPrefix(avatar, "/") {
		avatar = config.C.GlobalSteamDomain + avatar
	}
	return avatar
}

// Steam's generated avatar
func GetPlayerAvatar2(level int) string {

	ret := "avatar2"

	n100 := math.Floor(float64(level)/100) * 100
	if n100 >= 100 {

		ret += " lvl_" + FloatToString(n100, 0)

		n10 := math.Floor(float64(level)/10) * 10
		n10String := FloatToString(n10, 0)
		n10String = n10String[len(n10String)-2:]

		if n10String != "00" {
			ret += " lvl_plus_" + n10String
		}
	}

	return ret
}

func GetPlayerFlagPath(code string) string {

	if code == "" {
		return ""
	}
	return "/assets/img/flags/" + strings.ToLower(code) + ".png"
}

func GetPlayerPath(id int64, name string) string {

	p := "/players/" + strconv.FormatInt(id, 10)
	if name != "" {
		s := slug.Make(name)
		if s != "" {
			p = p + "/" + s
		}
	}
	return p
}

func GetPlayerPathAbsolute(id int64, name string) string {

	pathx := GetPlayerPath(id, name)

	if strings.HasPrefix(pathx, "/") {
		pathx = config.C.GlobalSteamDomain + pathx
	}

	return pathx
}

func GetPlayerName(id int64, name string) string {

	name = RegexFilterEmptyCharacters.ReplaceAllString(name, "")
	name = strings.TrimSpace(name)

	if name != "" {
		return name
	} else if id > 0 {
		return "-no name-" // No name
	} else {
		return "Unknown Player" // Name name/ID
	}
}

func IsValidPlayerID(id int64) (int64, error) {

	if id == 0 {
		return id, steamid.ErrInvalidPlayerID
	}

	s := strconv.FormatInt(id, 10)

	if !strings.HasPrefix(s, "765") {
		return id, steamid.ErrInvalidPlayerID
	}

	steamID, err := steamid.ParsePlayerID(s)
	if err != nil {
		return id, steamid.ErrInvalidPlayerID
	}

	return int64(steamID), nil
}

func GetPlayerMaxFriends(level int) (ret int) {
	ret = 250 + (level * 5)
	if ret > 2000 {
		return 2000
	}
	return ret
}

type RankMetric string

func (rk RankMetric) String() string {
	switch rk {
	case RankKeyLevel:
		return "Level"
	case RankKeyBadges:
		return "Badges"
	case RankKeyBadgesFoil:
		return "Foil Badges"
	case RankKeyGames:
		return "Games"
	case RankKeyAchievements:
		return "Achievements"
	case RankKeyPlaytime:
		return "Playtime"
	}
	return ""
}

func (rk RankMetric) Letter() string {
	return string(rk)
}

const (
	RankKeyLevel          RankMetric = "l"
	RankKeyBadges         RankMetric = "b"
	RankKeyBadgesFoil     RankMetric = "d"
	RankKeyGames          RankMetric = "g"
	RankKeyPlaytime       RankMetric = "p"
	RankKeyAchievements   RankMetric = "a"
	RankKeyAwardsGiven    RankMetric = "c"
	RankKeyAwardsReceived RankMetric = "e"
)

// Mongo col -> Rank key
var PlayerRankFields = map[string]RankMetric{
	"level":             RankKeyLevel,
	"games_count":       RankKeyGames,
	"badges_count":      RankKeyBadges,
	"badges_foil_count": RankKeyBadgesFoil,
	"play_time":         RankKeyPlaytime,
	"achievement_count": RankKeyAchievements,
	"awards_given":      RankKeyAwardsGiven,
	"awards_received":   RankKeyAwardsReceived,
}

// Rank key -> Influx col
var PlayerRankFieldsInflux = map[RankMetric]string{
	RankKeyLevel:          InfPlayersLevelRank.String(),
	RankKeyGames:          InfPlayersGamesRank.String(),
	RankKeyBadges:         InfPlayersBadgesRank.String(),
	RankKeyBadgesFoil:     InfPlayersBadgesFoilRank.String(),
	RankKeyPlaytime:       InfPlayersPlaytimeRank.String(),
	RankKeyAchievements:   InfPlayersAchievementsRank.String(),
	RankKeyAwardsGiven:    InfPlayersAwardsGiven.String(),
	RankKeyAwardsReceived: InfPlayersAwardsReceived.String(),
}
