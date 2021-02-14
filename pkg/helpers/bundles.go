package helpers

import (
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/i18n"
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
	GetDiscountSale() int
	GetDiscountHighest() int
	GetPrices() map[steamapi.ProductCC]int
	GetPricesFormatted() map[steamapi.ProductCC]string
	GetPricesSaleFormatted() map[steamapi.ProductCC]string
	GetScore() float64
	GetType() string
	GetApps() int
	GetPackages() int
	IsGiftable() bool
}

func OutputBundleForJSON(bundle Bundle) []interface{} {

	updated := strconv.FormatInt(bundle.GetUpdated().Unix(), 10)
	highest := bundle.GetDiscountHighest() == bundle.GetDiscount() && bundle.GetDiscount() != 0

	return []interface{}{
		bundle.GetID(),                  // 0
		bundle.GetName(),                // 1
		bundle.GetPath(),                // 2
		updated,                         // 3
		bundle.GetDiscount(),            // 4
		bundle.GetApps(),                // 5
		bundle.IsGiftable(),             // 6
		highest,                         // 7
		bundle.GetStoreLink(),           // 8
		bundle.GetPricesFormatted(),     // 9
		bundle.GetScore(),               // 10
		bundle.GetPricesSaleFormatted(), // 11
		bundle.GetType(),                // 12
		bundle.GetDiscountSale(),        // 13
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

func GetBundlePricesFormatted(prices map[steamapi.ProductCC]int) (ret map[steamapi.ProductCC]string) {

	ret = map[steamapi.ProductCC]string{}

	for k, v := range prices {
		ret[k] = i18n.FormatPrice(i18n.GetProdCC(k).CurrencyCode, v)
	}

	return ret
}
