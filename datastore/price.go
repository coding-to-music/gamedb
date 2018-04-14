package datastore

import (
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/gosimple/slug"
	"github.com/steam-authority/steam-authority/helpers"
)

var (
	cachePricesCount int
)

type Price struct {
	CreatedAt       time.Time `datastore:"created_at"`
	AppID           int       `datastore:"app_id"`
	PackageID       int       `datastore:"package_id"`
	Name            string    `datastore:"app_name"`
	PriceInitial    int       `datastore:"price_initial"`
	PriceFinal      int       `datastore:"price_final"`
	Discount        int       `datastore:"discount"`
	Currency        string    `datastore:"currency"`
	Change          int       `datastore:"change"`
	Icon            string    `datastore:"logo"`
	ReleaseDateNice string    `datastore:"release_date"`
	ReleaseDateUnix int64     `datastore:"release_date_unix"`
}

func (p Price) GetKey() (key *datastore.Key) {
	return datastore.IncompleteKey(KindPrice, nil)
}

func (p Price) GetPath() string {
	if p.AppID != 0 {
		return "/games/" + strconv.Itoa(p.AppID) + "/" + slug.Make(p.Name)
	} else if p.PackageID != 0 {
		return "/packages/" + strconv.Itoa(p.PackageID) + "/" + slug.Make(p.Name)
	} else {
		return ""
	}
}

func (p Price) GetIcon() (ret string) {

	if p.Icon == "" {
		return "/assets/img/no-app-image-square.jpg"
	} else if strings.HasPrefix(p.Icon, "/") {
		return p.Icon
	} else {
		return "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/" + strconv.Itoa(p.AppID) + "/" + p.Icon + ".jpg"
	}
}

func (p Price) GetCreatedNice() (ret string) {
	return p.CreatedAt.Format(helpers.DateTime)
}

func (p Price) GetCreatedUnix() (ret string) {
	return p.CreatedAt.Format(helpers.DateTime)
}

func (p Price) GetPriceInitial() float64 {
	return helpers.CentsInt(p.PriceInitial)
}

func (p Price) GetChange() float64 {
	return helpers.CentsInt(p.Change)
}

func (p Price) GetPriceFinal() float64 {
	return helpers.CentsInt(p.PriceFinal)
}

func GetAppPrices(appID int) (prices []Price, err error) {

	client, ctx, err := getClient()
	if err != nil {
		return prices, err
	}

	q := datastore.NewQuery(KindPrice).Order("created_at").Limit(100)
	q = q.Filter("app_id =", appID)
	q = q.Filter("currency =", "usd")

	_, err = client.GetAll(ctx, q, &prices)

	return prices, err
}

func GetPackagePrices(packageID int) (prices []Price, err error) {

	client, ctx, err := getClient()
	if err != nil {
		return prices, err
	}

	q := datastore.NewQuery(KindPrice).Order("created_at").Limit(100)
	q = q.Filter("package_id =", packageID)
	q = q.Filter("currency =", "usd")

	_, err = client.GetAll(ctx, q, &prices)

	return prices, err
}

func GetLatestPrices(limit int, page int) (prices []Price, err error) {

	client, ctx, err := getClient()
	if err != nil {
		return prices, err
	}

	offset := (page - 1) * limit

	q := datastore.NewQuery(KindPrice).Order("-created_at").Limit(limit).Offset(offset)
	q = q.Filter("currency =", "usd")

	_, err = client.GetAll(ctx, q, &prices)

	return prices, err
}

func CountPrices() (count int, err error) {

	if cachePricesCount == 0 {

		client, ctx, err := getClient()
		if err != nil {
			return count, err
		}

		q := datastore.NewQuery(KindPrice)
		cachePricesCount, err = client.Count(ctx, q)
		if err != nil {
			return count, err
		}
	}

	return cachePricesCount, nil
}
