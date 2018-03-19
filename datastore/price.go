package datastore

import (
	"time"

	"cloud.google.com/go/datastore"
)

type AppPrice struct {
	CreatedAt    time.Time `datastore:"created_at"`
	AppID        int       `datastore:"app_id"`
	PriceInitial int       `datastore:"price_initial"`
	PriceFinal   int       `datastore:"price_final"`
	Discount     int       `datastore:"discount"`
	Currency     string    `datastore:"currency"`
	Change       int       `datastore:"change"`
}

func (price AppPrice) GetKey() (key *datastore.Key) {
	return datastore.IncompleteKey(KindPriceApp, nil)
}

type PackagePrice struct {
	CreatedAt    time.Time `datastore:"created_at"`
	PackageID    int       `datastore:"package_id"`
	PriceInitial int       `datastore:"price_initial"`
	PriceFinal   int       `datastore:"price_final"`
	Discount     int       `datastore:"discount"`
	Currency     string    `datastore:"currency"`
	Change       int       `datastore:"change"`
}

func (pack PackagePrice) GetKey() (key *datastore.Key) {
	return datastore.IncompleteKey(KindPricePackage, nil)
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

func GetAppChanges() (prices []AppPrice, err error) {

	client, ctx, err := getDSClient()
	if err != nil {
		return prices, err
	}

	q := datastore.NewQuery(KindPriceApp).Order("created_at").Limit(1000)
	q = q.Filter("currency =", "usd")

	_, err = client.GetAll(ctx, q, &prices)

	return prices, err
}
