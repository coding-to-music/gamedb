package db

import (
	"encoding/json"
	"errors"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/logging"
)

type ProductType string

const (
	ProductTypeApp     ProductType = "product"
	ProductTypePackage ProductType = "package"
)

var ErrInvalidCountryCode = errors.New("invalid code")

type productInterface interface {
	GetID() int
	GetType() ProductType
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
	return price, ErrInvalidCountryCode
}

func (p ProductPrices) String() string {

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

func (p ProductPriceCache) GetInitial(code steam.CountryCode) string {
	return helpers.CurrencyFormat(code, p.Initial)
}

func (p ProductPriceCache) GetFinal(code steam.CountryCode) string {
	return helpers.CurrencyFormat(code, p.Final)
}

func (p ProductPriceCache) GetDiscountPercent() int {
	return p.DiscountPercent
}

func (p ProductPriceCache) GetIndividual(code steam.CountryCode) string {
	return helpers.CurrencyFormat(code, p.Individual)
}

func (p ProductPriceCache) GetCountryName(code steam.CountryCode) string {
	return helpers.CountryCodeToName(string(code))
}
