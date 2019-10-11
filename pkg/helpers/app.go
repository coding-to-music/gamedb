package helpers

import (
	"strconv"
	"strings"
	"time"

	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gosimple/slug"
)

type ProductType string

const (
	ProductTypeApp     ProductType = "product"
	ProductTypePackage ProductType = "package"

	DefaultAppIcon = "/assets/img/no-app-image-square.jpg"
)

func IsValidAppID(id int) bool {
	return id != 0
}

func GetAppPath(id int, name string) string {

	p := "/apps/" + strconv.Itoa(id)

	if name != "" {
		p = p + "/" + slug.Make(name)
	}

	return p
}

func GetAppName(id int, name string) string {

	if name != "" {
		return name
	} else if id > 0 {
		return "App " + strconv.Itoa(id)
	}
	return "Unknown App"
}

func GetAppIcon(id int, icon string) string {

	if icon == "" {
		return DefaultAppIcon
	} else if strings.HasPrefix(icon, "/") || strings.HasPrefix(icon, "http") {
		return icon
	}
	return "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/" + strconv.Itoa(id) + "/" + icon + ".jpg"
}

func GetAppReleaseState(state string) (ret string) {

	switch state {
	case "preloadonly":
		return "Preload Only"
	case "prerelease":
		return "Pre Release"
	case "released":
		return "Released"
	case "":
		return "Unreleased"
	default:
		log.Warning("Missing state: " + state)
		return strings.Title(state)
	}
}

func GetAppReleaseDateNice(releaseDateUnix int64, releaseDate string) string {

	if releaseDateUnix == 0 {
		return releaseDate
	}

	return time.Unix(releaseDateUnix, 0).Format(DateYear)
}
