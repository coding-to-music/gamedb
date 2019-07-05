package helpers

import (
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/gosimple/slug"
)

type ProductType string

const (
	ProductTypeApp     ProductType = "product"
	ProductTypePackage ProductType = "package"

	DefaultAppIcon = "https://gamedb.online/assets/img/no-app-image-square.jpg" // Absolute for Discord to hotlink
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
		return "App " + humanize.Comma(int64(id))
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
