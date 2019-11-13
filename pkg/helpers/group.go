package helpers

import (
	"math/big"
	"strings"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gosimple/slug"
)

const (
	AvatarBase = "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/"

	GroupTypeGame  = "game"
	GroupTypeGroup = "group"
)

func IsValidGroupID(id string) bool {

	if id == "" {
		return false
	}

	if len(id) > 8 && len(id) != 18 {
		return false
	}

	i := big.NewInt(0)
	i, success := i.SetString(id, 10)
	if !success || i == nil {
		return false
	}

	return true
}

func GetGroupPath(id64 string, name string) string {
	return "/groups/" + id64 + "/" + slug.Make(name)
}

func GetGroupType(typex string) string {
	return strings.Title(typex)
}

func IsGroupOfficial(typex string) bool {
	return typex == GroupTypeGame
}

func GetGroupLink(typex string, url string, ) string {
	return "https://steamcommunity.com/" + typex + "s/" + url + "?utm_source=" + config.Config.GameDBShortName.Get()
}

func GetGroupName(name string, id64 string) string {
	if name == "" {
		return "Group " + id64
	}
	return name
}

func GetGroupIcon(icon string) string {
	return AvatarBase + icon
}
