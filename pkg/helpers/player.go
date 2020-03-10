package helpers

import (
	"errors"
	"math"
	"strconv"
	"strings"

	"github.com/Jleagle/steam-go/steamid"
	"github.com/gosimple/slug"
)

const (
	DefaultPlayerAvatar = "/assets/img/no-player-image.jpg"
)

var (
	ErrInvalidPlayerID = errors.New("invalid player id")
)

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

func IsValidPlayerID(id int64) (int64, error) {

	if id == 0 {
		return id, ErrInvalidPlayerID
	}

	s := strconv.FormatInt(id, 10)

	if !strings.HasPrefix(s, "7656") {
		return id, ErrInvalidPlayerID
	}

	if len(s) != 17 {
		return id, ErrInvalidPlayerID
	}

	steamID, err := steamid.ParsePlayerID(s)
	if err != nil {
		return id, ErrInvalidGroupID
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
