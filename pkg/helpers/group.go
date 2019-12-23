package helpers

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Philipp15b/go-steam/steamid"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gosimple/slug"
)

const (
	AvatarBase = "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/"

	GroupTypeGame  = "game"
	GroupTypeGroup = "group"
)

func IsValidGroupID(id string) bool {

	// if len(id) != 18 {
	// 	return false
	// }

	if !RegexNumbers.MatchString(id) {
		return false
	}

	return true
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
	if name == "" {
		return "Group " + id
	}
	return name
}

func GetGroupIcon(icon string) string {
	return AvatarBase + icon
}

var ErrInvalidGroupID = errors.New("invalid group id")

func UpgradeGroupID(id string) (string, error) {

	if len(id) > 18 {

		return id, ErrInvalidGroupID

	} else if len(id) < 18 {

		i, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			return id, err
		}

		steamID := steamid.NewIdAdv(uint32(i), 0, 1, 7)
		id = strconv.FormatUint(uint64(steamID), 10)
	}

	return id, nil
}
