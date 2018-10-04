package db

import (
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/gosimple/slug"
	"github.com/steam-authority/steam-authority/helpers"
)

type AppPrice struct {
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
	First           bool      `datastore:"first"`
}

func (p AppPrice) GetKey() (key *datastore.Key) {
	return datastore.IncompleteKey(KindPrice, nil)
}

func (p AppPrice) GetPath() string {
	if p.AppID != 0 {
		return "/games/" + strconv.Itoa(p.AppID) + "/" + slug.Make(p.Name)
	} else if p.PackageID != 0 {
		return "/packages/" + strconv.Itoa(p.PackageID) + "/" + slug.Make(p.Name)
	} else {
		return ""
	}
}

func (p AppPrice) GetIcon() (ret string) {

	if p.Icon == "" {
		return "/assets/img/no-app-image-square.jpg"
	} else if strings.HasPrefix(p.Icon, "/") {
		return p.Icon
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

func (p AppPrice) GetPriceInitial() float64 {
	return helpers.CentsInt(p.PriceInitial)
}

func (p AppPrice) GetChange() float64 {
	return helpers.CentsInt(p.Change)
}

func (p AppPrice) GetPriceFinal() float64 {
	return helpers.CentsInt(p.PriceFinal)
}

//func (p AppPrice) GetChangePercent() float64 {
//
//	if p.Change < 0 {
//		// Green
//		old := p.PriceFinal + p.Change
//		return helpers.CentsInt(old / p.Change)
//	} else {
//		// Red
//		old := p.PriceFinal + p.Change
//		return helpers.CentsInt(old / p.Change)
//	}
//}

// Data array for datatables
func (p AppPrice) OutputForJSON() (output []interface{}) {

	return []interface{}{
		p.AppID,
		p.Name,
		p.ReleaseDateNice,
		p.GetPriceFinal(),
		p.Discount,
		p.GetChange(),
		p.GetCreatedNice(),
		p.GetPath(),
		p.GetIcon(),
	}
}

func GetAppPrices(appID int, limit int) (prices []AppPrice, err error) {

	client, ctx, err := GetDSClient()
	if err != nil {
		return prices, err
	}

	if limit == 0 {
		limit = 100
	}

	q := datastore.NewQuery(KindPrice).Order("created_at").Limit(limit)
	q = q.Filter("app_id =", appID)
	q = q.Filter("currency =", "usd")

	_, err = client.GetAll(ctx, q, &prices)
	if err != nil {
		return
	}

	return prices, err
}

func GetPackagePrices(packageID int, limit int) (prices []AppPrice, err error) {

	client, ctx, err := GetDSClient()
	if err != nil {
		return prices, err
	}

	if limit == 0 {
		limit = 100
	}

	q := datastore.NewQuery(KindPrice).Order("created_at").Limit(limit)
	q = q.Filter("package_id =", packageID)
	q = q.Filter("currency =", "usd")

	_, err = client.GetAll(ctx, q, &prices)
	if err != nil {
		return
	}

	return prices, err
}
