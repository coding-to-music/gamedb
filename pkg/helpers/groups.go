package helpers

import (
	"errors"
	"strings"

	"github.com/Jleagle/steam-go/steamid"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gosimple/slug"
)

const (
	AvatarBase = "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/"

	GroupTypeGame  = "game"
	GroupTypeGroup = "group"
)

var ErrInvalidGroupID = errors.New("invalid group id")

func IsValidGroupID(id string) (string, error) {

	if id == "" {
		return id, ErrInvalidGroupID
	}

	steamID, err := steamid.ParseGroupID(id)
	if err != nil {
		return id, ErrInvalidGroupID
	}

	id = steamID.String()

	if !strings.HasPrefix(id, "1035") {
		return id, ErrInvalidGroupID
	}

	return id, nil
}

func GetGroupPath(id string, name string) string {
	return "/groups/" + id + "/" + slug.Make(name)
}

func GetGroupPathAbsolute(id string, name string) string {

	pathx := GetGroupPath(id, name)

	if strings.HasPrefix(pathx, "/") {
		pathx = config.C.GameDBDomain + pathx
	}

	return pathx
}

func GetGroupType(typex string) string {
	return strings.Title(typex)
}

func IsGroupOfficial(typex string) bool {
	return typex == GroupTypeGame
}

func GetGroupLink(typex string, url string, ) string {
	return "https://steamcommunity.com/" + typex + "s/" + url + "?utm_source=" + config.C.GameDBShortName
}

func GetGroupName(id string, name string) string {

	name = RegexFilterEmptyCharacters.ReplaceAllString(name, "")
	name = strings.TrimSpace(name)

	if name == "" {
		return "-no name-"
	}
	return name
}

func GetGroupAbbreviation(name string) string {

	name = RegexFilterEmptyCharacters.ReplaceAllString(name, "")
	name = strings.TrimSpace(name)

	return name
}

func GetGroupIcon(icon string) string {

	if strings.HasPrefix(icon, "/") || strings.HasPrefix(icon, "http") {
		return icon
	}

	if icon == "" {
		return DefaultPlayerAvatar
	}

	return AvatarBase + icon
}
