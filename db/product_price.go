package db

import (
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/helpers"
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
	return datastore.IncompleteKey(KindAppPrice, nil)
}

func (p ProductPrice) GetPath() string {
	if p.AppID != 0 {
		return getAppPath(p.AppID, p.Name)
	} else if p.PackageID != 0 {
		return getPackagePath(p.PackageID, p.Name)
	} else {
		return ""
	}
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

func (p ProductPrice) GetCreatedNice() string {
	return p.CreatedAt.Format(helpers.DateTime)
}

func (p ProductPrice) GetCreatedUnix() int64 {
	return p.CreatedAt.Unix()
}

func (p ProductPrice) GetPriceBefore() float64 {
	return helpers.CentsInt(p.PriceBefore)
}

func (p ProductPrice) GetPriceAfter() float64 {
	return helpers.CentsInt(p.PriceAfter)
}

func (p ProductPrice) GetDifference() string {

	diff := strconv.FormatFloat(float64(p.Difference)/100, 'f', 2, 64)

	if p.Difference > 0 {
		return "+" + diff
	} else if p.Difference < 0 {
		return diff
	}

	return "0"
}

func (p ProductPrice) GetDifferencePercent() string {

	diff := strconv.FormatFloat(p.DifferencePercent, 'f', 2, 64)

	if p.DifferencePercent > 0 {
		return "+" + diff + "%"
	} else if p.Difference < 0 {
		return "-" + diff + "%"
	}

	return "0%"
}

// Data array for datatables
func (p ProductPrice) OutputForJSON() (output []interface{}) {

	return []interface{}{
		p.AppID,
		p.PackageID,
		p.Currency,
		p.Name,
		p.GetIcon(),
		p.GetPath(),
		p.GetPriceBefore(),
		p.GetPriceAfter(),
		p.GetDifference(),
		p.GetDifferencePercent(),
		p.GetCreatedNice(),
		p.GetCreatedUnix(),
	}
}

// Does not save here so can be bulk saved elsewear
func CreateProductPrice(product productInterface, currency steam.CountryCode, priceBefore int, priceAfter int) ProductPrice {

	price := ProductPrice{}

	if product.GetType() == ProductTypeApp {
		price.AppID = product.GetID()
	} else if product.GetType() == ProductTypeApp {
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

func GetProductPrices(ID int, productType productType) (prices []ProductPrice, err error) {

	client, ctx, err := GetDSClient()
	if err != nil {
		return prices, err
	}

	q := datastore.NewQuery(KindAppPrice).Order("created_at").Limit(1000)

	if productType == ProductTypeApp {
		q = q.Filter("app_id =", ID)
	} else if productType == ProductTypePackage {
		q = q.Filter("package_id =", ID)
	} else {
		return prices, err
	}

	_, err = client.GetAll(ctx, q, &prices)
	return prices, err
}
