package helpers

import (
	"math"
	"strconv"
	"strings"

	"github.com/Jleagle/steam-go/steamid"
	"github.com/gosimple/slug"
)

const (
	DefaultPlayerAvatar = "/assets/img/no-player-image.jpg"
)

func GetPlayerCommunityLink(playerID int64, vanityURL string) string {

	// Just use ID in case url has changed
	// if vanityURL != "" && vanityURL != strconv.FormatInt(playerID, 10) {
	// 	return "https://steamcommunity.com/id/" + vanityURL
	// }

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
