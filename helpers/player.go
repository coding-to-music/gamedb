package helpers

import (
	"strconv"
	"strings"

	"github.com/gosimple/slug"
)

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
