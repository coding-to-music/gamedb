package db

import (
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
)

type ProductPrice struct {
	CreatedAt         time.Time         `datastore:"created_at,noindex"`
	AppID             int               `datastore:"app_id,noindex"`
	PackageID         int               `datastore:"package_id,noindex"`
	Currency          steam.CountryCode `datastore:"currency,noindex"`
	Name              string            `datastore:"name,noindex"`
	Icon              string            `datastore:"icon,noindex"`
	PriceBefore       int               `datastore:"price_before,noindex"` // cents
	PriceAfter        int               `datastore:"price_after,noindex"`  // cents
	Difference        int               `datastore:"difference,noindex"`
	DifferencePercent float64           `datastore:"difference_percent,noindex"`
}

func (p ProductPrice) GetKey() (key *datastore.Key) {
	return datastore.IncompleteKey(KindProductPrice, nil)
}

func (p ProductPrice) GetPath() string {

	if p.AppID != 0 {
		return helpers.GetAppPath(p.AppID, p.Name)
	} else if p.PackageID != 0 {
		return GetPackagePath(p.PackageID, p.Name)
	}

	return ""
}

func (p ProductPrice) GetIcon() string {
	if p.Icon == "" {
		return helpers.DefaultAppIcon
	} else if strings.HasPrefix(p.Icon, "/") || strings.HasPrefix(p.Icon, "http") {
		return p.Icon
	} else {
		return "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/" + strconv.Itoa(p.AppID) + "/" + p.Icon + ".jpg"
	}
}

func (p ProductPrice) GetPercentChange() float64 {

	if math.IsInf(p.DifferencePercent, 0) {
		return 0
	}
	return helpers.RoundFloatTo2DP(p.DifferencePercent)

}

// Data array for datatables
func (p ProductPrice) OutputForJSON() (output []interface{}) {

	locale, err := helpers.GetLocaleFromCountry(p.Currency)
	log.Err(err)

	return []interface{}{
		p.AppID,                              // 0
		p.PackageID,                          // 1
		p.Currency,                           // 2
		p.Name,                               // 3
		p.GetIcon(),                          // 4
		p.GetPath(),                          // 5
		locale.Format(p.PriceBefore),         // 6
		locale.Format(p.PriceAfter),          // 7
		locale.Format(p.Difference),          // 8
		p.GetPercentChange(),                 // 9
		p.CreatedAt.Format(helpers.DateTime), // 10
		p.CreatedAt.Unix(),                   // 11
		p.Difference,                         // 12 Raw difference
	}
}

// Does not save here so can be bulk saved elsewear
func CreateProductPrice(product ProductInterface, currency steam.CountryCode, priceBefore int, priceAfter int) ProductPrice {

	price := ProductPrice{}

	if product.GetProductType() == helpers.ProductTypeApp {
		price.AppID = product.GetID()
	} else if product.GetProductType() == helpers.ProductTypePackage {
		price.PackageID = product.GetID()
	} else {
		panic("Invalid productType")
	}

	price.Name = product.GetName()
	price.Icon = product.GetIcon()
	price.CreatedAt = time.Now()
	price.Currency = currency
	price.PriceBefore = priceBefore
	price.PriceAfter = priceAfter
	price.Difference = priceAfter - priceBefore
	price.DifferencePercent = (float64(priceAfter-priceBefore) / float64(priceBefore)) * 100

	return price
}

func GetProductPrices(ID int, productType helpers.ProductType, currency steam.CountryCode) (prices []ProductPrice, err error) {

	client, ctx, err := GetDSClient()
	if err != nil {
		return prices, err
	}

	q := datastore.NewQuery(KindProductPrice).Order("-created_at").Limit(1000)
	q = q.Filter("currency =", string(currency))

	if productType == helpers.ProductTypeApp {
		q = q.Filter("app_id =", ID)
	} else if productType == helpers.ProductTypePackage {
		q = q.Filter("package_id =", ID)
	}

	_, err = client.GetAll(ctx, q, &prices)

	// Reverse order for frontend
	sort.Slice(prices, func(i, j int) bool {
		return prices[i].CreatedAt.Unix() > prices[j].CreatedAt.Unix()
	})

	return prices, err
}

func ChunkPrice(kinds []ProductPrice) (chunked [][]ProductPrice) {

	for i := 0; i < len(kinds); i += 500 {
		end := i + 500

		if end > len(kinds) {
			end = len(kinds)
		}

		chunked = append(chunked, kinds[i:end])
	}

	return chunked
}
