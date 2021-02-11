package helpers

import (
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gosimple/slug"
)

type Bundle interface {
	OutputForJSON() (output []interface{})
	GetName() string
	GetPath() string
	GetStoreLink() string
	GetID() int
	GetUpdated() time.Time
	GetDiscount() int
	GetDiscountHighest() int
	GetPrices() map[steamapi.ProductCC]int
	GetScore() float64
	GetApps() int
	GetPackages() int
}

func OutputBundleForJSON(bundle Bundle) []interface{} {

	updated := strconv.FormatInt(bundle.GetUpdated().Unix(), 10)
	highest := bundle.GetDiscountHighest() == bundle.GetDiscount() && bundle.GetDiscount() != 0

	return []interface{}{
		bundle.GetID(),        // 0
		bundle.GetName(),      // 1
		bundle.GetPath(),      // 2
		updated,               // 3
		bundle.GetDiscount(),  // 4
		bundle.GetApps(),      // 5
		bundle.GetPackages(),  // 6
		highest,               // 7
		bundle.GetStoreLink(), // 8
		bundle.GetPrices(),    // 9
		bundle.GetScore(),     // 10
	}
}

func GetBundlePath(id int, name string) string {
	return "/bundles/" + strconv.Itoa(id) + "/" + slug.Make(GetBundleName(id, name))
}

func GetBundleName(id int, name string) string {

	name = strings.TrimSpace(name)

	if name != "" {
		return name
	}

	return "Bundle " + strconv.Itoa(id)
}

func GetBundleStoreLink(id int) string {
	return "https://store.steampowered.com/bundle/" + strconv.Itoa(id) +
		"?utm_source=" + config.C.GameDBShortName + "&utm_medium=referral&utm_campaign=app-store-link"
}
