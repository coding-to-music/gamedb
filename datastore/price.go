package datastore

import (
	"strconv"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/steam-authority/steam-authority/helpers"
)

var (
	cachePricesCount  int
)

type AppPrice struct {
	CreatedAt       time.Time `datastore:"created_at"`
	AppID           int       `datastore:"app_id"`
	AppName         string    `datastore:"app_name"`
	PriceInitial    int       `datastore:"price_initial"`
	PriceFinal      int       `datastore:"price_final"`
	Discount        int       `datastore:"discount"`
	Currency        string    `datastore:"currency"`
	Change          int       `datastore:"change"`
	Icon            string    `datastore:"logo"`
	ReleaseDateNice string    `datastore:"release_date"`
	ReleaseDateUnix int64     `datastore:"release_date_unix"`
}

func (p AppPrice) GetKey() (key *datastore.Key) {
	return datastore.IncompleteKey(KindPriceApp, nil)
}

func (p AppPrice) GetLogo() (ret string) {

	if p.Icon == "" {
		return "/assets/img/no-app-image-square.jpg"
	} else {
		return "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/" + strconv.Itoa(p.AppID) + "/" + p.Icon + ".jpg"
	}
}

func (p AppPrice) GetCreatedNice() (ret string) {
	return p.CreatedAt.Format(helpers.DateTime)
}

func (p AppPrice) GetCreatedUnix() (ret string) {
	return p.CreatedAt.Format(helpers.DateTime)
}

func (p AppPrice) GetPriceInitial() string {
	return helpers.CentsInt(p.PriceInitial)
}

func (p AppPrice) GetChange() string {
	return helpers.CentsInt(p.Change)
}

func (p AppPrice) GetPriceFinal() string {
	return helpers.CentsInt(p.PriceFinal)
}

func GetAppPrices(appID int) (prices []AppPrice, err error) {

	client, ctx, err := getClient()
	if err != nil {
		return prices, err
	}

	q := datastore.NewQuery(KindPriceApp).Order("created_at").Limit(1000)
	q = q.Filter("app_id =", appID)
	q = q.Filter("currency =", "usd")

	_, err = client.GetAll(ctx, q, &prices)

	return prices, err
}

func GetAppChanges(limit int, page int) (prices []AppPrice, err error) {

	client, ctx, err := getClient()
	if err != nil {
		return prices, err
	}

	offset := (page - 1) * limit

	q := datastore.NewQuery(KindPriceApp).Order("-created_at").Limit(limit).Offset(offset)
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

		q := datastore.NewQuery(KindPriceApp)
		cachePricesCount, err = client.Count(ctx, q)
		if err != nil {
			return count, err
		}
	}

	return cachePricesCount, nil
}

type PackagePrice struct {
	CreatedAt    time.Time `datastore:"created_at"`
	PackageID    int       `datastore:"package_id"`
	PriceInitial int       `datastore:"price_initial"`
	PriceFinal   int       `datastore:"price_final"`
	Discount     int       `datastore:"discount"`
	Currency     string    `datastore:"currency"`
	Change       int       `datastore:"change"`
	Logo         string    `datastore:"logo"`
}

func (p PackagePrice) GetKey() (key *datastore.Key) {
	return datastore.IncompleteKey(KindPricePackage, nil)
}

func GetPackagePrices(packageID int) (prices []PackagePrice, err error) {

	client, ctx, err := getClient()
	if err != nil {
		return prices, err
	}

	q := datastore.NewQuery(KindPricePackage).Order("created_at").Limit(1000)
	q = q.Filter("package_id =", packageID)
	q = q.Filter("currency =", "usd")

	_, err = client.GetAll(ctx, q, &prices)

	return prices, err
}
