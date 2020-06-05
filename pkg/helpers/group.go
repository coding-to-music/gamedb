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

func GetGroupType(typex string) string {
	return strings.Title(typex)
}

func IsGroupOfficial(typex string) bool {
	return typex == GroupTypeGame
}

func GetGroupLink(typex string, url string, ) string {
	return "https://steamcommunity.com/" + typex + "s/" + url + "?utm_source=" + config.Config.GameDBShortName.Get()
}

func GetGroupName(name string, id string) string {

	name = RegexFilterEmptyCharacters.ReplaceAllString(name, "")
	name = strings.TrimSpace(name)

	if name == "" {
		return "Group " + id
	}
	return name
}

func GetGroupAbbreviation(name string) string {

	name = RegexFilterEmptyCharacters.ReplaceAllString(name, "")
	name = strings.TrimSpace(name)

	return name
}

func GetGroupIcon(icon string) string {

	if icon == "" {
		return "/assets/img/no-app-image-square.jpg"
	}
	return AvatarBase + icon
}
