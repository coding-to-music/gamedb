package db

import (
	"encoding/json"
	"errors"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/logging"
)

type productType string

const (
	ProductTypeApp     productType = "product"
	ProductTypePackage             = "package"
)

type productInterface interface {
	GetID() int
	GetType() productType
	GetName() string
	GetIcon() string
}

type ProductPrices map[steam.CountryCode]ProductPriceCache

func (p *ProductPrices) AddPriceFromPackage(code steam.CountryCode, prices steam.PackageDetailsBody) {

	(*p)[code] = ProductPriceCache{
		Currency:        prices.Data.Price.Currency,
		Initial:         prices.Data.Price.Initial,
		Final:           prices.Data.Price.Final,
		DiscountPercent: prices.Data.Price.DiscountPercent,
		Individual:      prices.Data.Price.Individual,
	}
}

func (p *ProductPrices) AddPriceFromApp(code steam.CountryCode, prices steam.AppDetailsBody) {
	(*p)[code] = ProductPriceCache{
		Currency:        prices.Data.PriceOverview.Currency,
		Initial:         prices.Data.PriceOverview.Initial,
		Final:           prices.Data.PriceOverview.Final,
		DiscountPercent: prices.Data.PriceOverview.DiscountPercent,
	}
}

func (p ProductPrices) Get(code steam.CountryCode) (price ProductPriceCache, err error) {
	if val, ok := p[code]; ok {
		return val, err
	}
	return price, errors.New("invalid code")
}

func (p ProductPrices) ToString() string {

	bytes, err := json.Marshal(p)
	logging.Error(err)
	return string(bytes)
}

type ProductPriceCache struct {
	Currency        string `json:"currency"`
	Initial         int    `json:"initial"`
	Final           int    `json:"final"`
	DiscountPercent int    `json:"discount_percent"`
	Individual      int    `json:"individual"`
}

func (p ProductPriceCache) GetInitial() float64 {
	return helpers.CentsInt(p.Initial)
}

func (p ProductPriceCache) GetFinal() float64 {
	return helpers.CentsInt(p.Final)
}

func (p ProductPriceCache) GetDiscountPercent() float64 {
	return helpers.CentsInt(p.Initial)
}

func (p ProductPriceCache) GetIndividual() float64 {
	return helpers.CentsInt(p.Final)
}
