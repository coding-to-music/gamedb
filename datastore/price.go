package datastore

import (
	"fmt"
	"strconv"
	"time"

	"cloud.google.com/go/datastore"
)

type AppPrice struct {
	CreatedAt    time.Time `datastore:"created_at"`
	AppID        int       `datastore:"app_id"`
	AppName      string    `datastore:"app_name"`
	PriceInitial int       `datastore:"price_initial"`
	PriceFinal   int       `datastore:"price_final"`
	Discount     int       `datastore:"discount"`
	Currency     string    `datastore:"currency"`
	Change       int       `datastore:"change"`
	Logo         string    `datastore:"logo"`
	ReleaseDate  string    `datastore:"release_date"`
}

func (p AppPrice) GetKey() (key *datastore.Key) {
	return datastore.IncompleteKey(KindPriceApp, nil)
}

func (p AppPrice) GetLogo() (ret string) {

	if p.Logo == "" {
		return "/assets/img/no-app-image-square.jpg"
	} else {
		return "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/apps/" + strconv.Itoa(p.AppID) + "/" + p.Logo + ".jpg"
	}
}

func (p AppPrice) GetPriceInitial() string {
	return fmt.Sprintf("%0.2f", float64(p.PriceInitial)/100)
}

func GetAppPrices(appID int) (prices []AppPrice, err error) {

	client, ctx, err := getDSClient()
	if err != nil {
		return prices, err
	}

	q := datastore.NewQuery(KindPriceApp).Order("created_at").Limit(1000)
	q = q.Filter("app_id =", appID)
	q = q.Filter("currency =", "usd")

	_, err = client.GetAll(ctx, q, &prices)

	return prices, err
}

func GetAppChanges() (prices []AppPrice, err error) {

	client, ctx, err := getDSClient()
	if err != nil {
		return prices, err
	}

	q := datastore.NewQuery(KindPriceApp).Order("-created_at").Limit(1000)
	q = q.Filter("currency =", "usd")

	_, err = client.GetAll(ctx, q, &prices)

	return prices, err
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

	client, ctx, err := getDSClient()
	if err != nil {
		return prices, err
	}

	q := datastore.NewQuery(KindPricePackage).Order("created_at").Limit(1000)
	q = q.Filter("package_id =", packageID)
	q = q.Filter("currency =", "usd")

	_, err = client.GetAll(ctx, q, &prices)

	return prices, err
}
