package helpers

import (
	"strconv"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/helpers/i18n"
)

type ProductInterface interface {
	GetID() int
	GetProductType() ProductType
	GetName() string
	GetIcon() string
	GetPrices() (prices ProductPrices)
	GetPath() string
	GetType() string
}

//
type ProductPrices map[steamapi.ProductCC]ProductPrice

func (p *ProductPrices) AddPriceFromPackage(code steamapi.ProductCC, prices steamapi.PackageDetailsBody) {

	if prices.Data.Price.Currency == "" {
		prices.Data.Price.Currency = i18n.GetProdCC(code).CurrencyCode
	}

	(*p)[code] = ProductPrice{
		Currency:        prices.Data.Price.Currency,
		Initial:         prices.Data.Price.Initial,
		Final:           prices.Data.Price.Final,
		DiscountPercent: prices.Data.Price.DiscountPercent,
		Individual:      prices.Data.Price.Individual,
	}
}

func (p *ProductPrices) AddPriceFromApp(code steamapi.ProductCC, prices steamapi.AppDetailsBody) {

	if prices.Data.PriceOverview.Currency == "" {
		prices.Data.PriceOverview.Currency = i18n.GetProdCC(code).CurrencyCode
	}

	(*p)[code] = ProductPrice{
		Currency:        prices.Data.PriceOverview.Currency,
		Initial:         prices.Data.PriceOverview.Initial,
		Final:           prices.Data.PriceOverview.Final,
		DiscountPercent: prices.Data.PriceOverview.DiscountPercent,
	}
}

func (p ProductPrices) Get(code steamapi.ProductCC) (price ProductPrice) {

	if val, ok := p[code]; ok {
		val.Exists = true
		return val
	}

	return price
}

func (p ProductPrices) Map() (prices map[steamapi.ProductCC]int) {

	prices = map[steamapi.ProductCC]int{}

	for k, v := range p {
		prices[k] = v.Final
	}

	return prices
}

//
type ProductPrice struct {
	Exists          bool                  `json:"-"`
	Currency        steamapi.CurrencyCode `json:"currency"`
	Initial         int                   `json:"initial"`
	Final           int                   `json:"final"`
	DiscountPercent int                   `json:"discount_percent"`
	Individual      int                   `json:"individual"`
}

func (p ProductPrice) GetDiscountPercent() string {
	return strconv.Itoa(p.DiscountPercent) + "%"
}

func (p ProductPrice) GetCountryName(code steamapi.ProductCC) string {
	return i18n.GetProdCC(code).Name
}

func (p ProductPrice) GetFlag(code steamapi.ProductCC) string {
	return "/assets/img/flags/" + i18n.GetProdCC(code).GetFlag() + ".png"
}

func (p ProductPrice) GetInitial() string {
	return p.format(p.Initial)
}

func (p ProductPrice) GetFinal() string {
	return p.format(p.Final)
}

func (p ProductPrice) GetIndividual() string {
	return p.format(p.Individual)
}

func (p ProductPrice) format(value int) string {
	if p.Currency == "" || !p.Exists {
		return "-"
	}
	return i18n.FormatPrice(p.Currency, value)
}
