package helpers

import (
	"math"
	"strconv"
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/i18n"
	"github.com/jinzhu/now"
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

func (p *ProductPrices) AddPriceFromApp(code steamapi.ProductCC, prices steamapi.AppDetails) {

	if prices.Data == nil {
		return
	}

	pp := ProductPrice{
		Free: prices.Data.IsFree,
	}

	if prices.Data.PriceOverview != nil {

		if prices.Data.PriceOverview.Currency == "" {
			prices.Data.PriceOverview.Currency = i18n.GetProdCC(code).CurrencyCode
		}

		pp.Currency = prices.Data.PriceOverview.Currency
		pp.Initial = prices.Data.PriceOverview.Initial
		pp.Final = prices.Data.PriceOverview.Final
		pp.DiscountPercent = prices.Data.PriceOverview.DiscountPercent
	}

	(*p)[code] = pp
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
		if v.Exists {
			prices[k] = v.Final
		}
	}

	return prices
}

//
type ProductPrice struct {
	Exists          bool                  `json:"-" bson:"-"`
	Currency        steamapi.CurrencyCode `json:"currency"`
	Initial         int                   `json:"initial"`
	Final           int                   `json:"final"`
	DiscountPercent int                   `json:"discount_percent"`
	Individual      int                   `json:"individual"`
	Free            bool                  `json:"free"`
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
	if p.Free && value == 0 {
		return "Free"
	}
	if !p.Exists || p.Currency == "" {
		return "-"
	}
	return i18n.FormatPrice(p.Currency, value)
}

var releaseDateFormats = []string{
	"2 Jan 2006",
	"2 Jan, 2006",
	"Jan 2, 2006",
	"Jan 2006",
	"January 2, 2006",
	"January 2006",
	// "2006", // Too wide a range
}

func GetReleaseDateUnix(date string) int64 {

	// for k, v := range map[string]string{"Q1 ": "January ", "Q2 ": "April ", "Q3 ": "July ", "Q4 ": "October "} {
	// 	if strings.HasPrefix(date, k) {
	// 		date = strings.Replace(date, k, v, 1)
	// 	}
	// }

	if date != "" {
		for _, v := range releaseDateFormats {
			t, err := time.Parse(v, date)
			if err == nil {
				return t.Unix()
			}
		}
	}

	return 0
}

func GetDaysToRelease(unix int64) string {

	release := time.Unix(unix, 0)

	days := math.Floor(release.Sub(now.BeginningOfDay()).Hours() / 24)

	if days == 0 {
		return "Today"
	}

	return "In " + GetTimeLong(int(days)*24*60, 2)
}
