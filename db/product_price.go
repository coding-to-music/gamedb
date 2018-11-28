package db

import (
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
	CreatedAt         time.Time         `datastore:"created_at"`
	AppID             int               `datastore:"app_id"`
	PackageID         int               `datastore:"package_id"`
	Currency          steam.CountryCode `datastore:"currency"`
	Name              string            `datastore:"name,noindex"`
	Icon              string            `datastore:"icon,noindex"`
	PriceBefore       int               `datastore:"price_before"` // cents
	PriceAfter        int               `datastore:"price_after"`  // cents
	Difference        int               `datastore:"difference"`
	DifferencePercent float64           `datastore:"difference_percent"`
}

func (p ProductPrice) GetKey() (key *datastore.Key) {
	return datastore.IncompleteKey(KindProductPrice, nil)
}

func (p ProductPrice) GetPath() string {

	if p.AppID != 0 {
		return getAppPath(p.AppID, p.Name)
	} else if p.PackageID != 0 {
		return getPackagePath(p.PackageID, p.Name)
	}

	return ""
}

func (p ProductPrice) GetIcon() string {

	if p.Icon == "" {
		return DefaultAppIcon
	} else if strings.HasPrefix(p.Icon, "/") {
		return p.Icon
	} else {
		return "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/" + strconv.Itoa(p.AppID) + "/" + p.Icon + ".jpg"
	}
}

func (p ProductPrice) GetPercentChange() string {

	return helpers.FloatToString(p.DifferencePercent, 0) + "%"

}

// Data array for datatables
func (p ProductPrice) OutputForJSON() (output []interface{}) {

	locale, err := helpers.GetLocaleFromCountry(p.Currency)
	log.Log(err)

	return []interface{}{
		p.AppID,
		p.PackageID,
		p.Currency,
		p.Name,
		p.GetIcon(),
		p.GetPath(),
		locale.Format(p.PriceBefore),
		locale.Format(p.PriceAfter),
		locale.Format(p.Difference),
		p.GetPercentChange(),
		p.CreatedAt.Format(helpers.DateTime),
		p.CreatedAt.Unix(),
		p.Difference, // Raw difference
	}
}

// Does not save here so can be bulk saved elsewear
func CreateProductPrice(product ProductInterface, currency steam.CountryCode, priceBefore int, priceAfter int) ProductPrice {

	price := ProductPrice{}

	if product.GetProductType() == ProductTypeApp {
		price.AppID = product.GetID()
	} else if product.GetProductType() == ProductTypePackage {
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

func GetProductPrices(ID int, productType ProductType, currency steam.CountryCode) (prices []ProductPrice, err error) {

	client, ctx, err := GetDSClient()
	if err != nil {
		return prices, err
	}

	q := datastore.NewQuery(KindProductPrice).Order("-created_at").Limit(1000)
	q = q.Filter("currency =", string(currency))

	if productType == ProductTypeApp {
		q = q.Filter("app_id =", ID)
	} else if productType == ProductTypePackage {
		q = q.Filter("package_id =", ID)
	}

	_, err = client.GetAll(ctx, q, &prices)

	// Reverse order for frontend
	sort.Slice(prices, func(i, j int) bool {
		return prices[i].CreatedAt.Unix() > prices[j].CreatedAt.Unix()
	})

	return prices, err
}
