package helpers

import (
	"math"
	"strconv"
	"strings"

	"github.com/gosimple/slug"
)

const DefaultPlayerAvatar = "/assets/img/no-player-image.jpg"

func GetPlayerAvatar(avatar string) string {

	if strings.HasPrefix(avatar, "http") || strings.HasPrefix(avatar, "/") {
		return avatar
	} else if avatar != "" {
		return AvatarBase + avatar
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
	name = strings.TrimSpace(name)
	if name != "" {
		return name
	} else if id > 0 {
		return "Player " + strconv.FormatInt(id, 10)
	} else {
		return "Unknown Player"
	}
}

func IsValidPlayerID(id int64) bool {

	if id == 0 {
		return false
	}

	idString := strconv.FormatInt(id, 10)

	if !strings.HasPrefix(idString, "76") {
		return false
	}

	if len(idString) != 17 {
		return false
	}

	return true
}

func GetPlayerMaxFriends(level int) (ret int) {

	ret = 750

	if level > 100 {
		ret = 1250
	}

	if level > 200 {
		ret = 1750
	}

	if level > 300 {
		ret = 2000
	}

	return ret
}
