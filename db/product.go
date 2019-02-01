package db

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
)

type ProductInterface interface {
	GetID() int
	GetProductType() ProductType
	GetName() string
	GetIcon() string
	GetPrice(code steam.CountryCode) (price ProductPriceStruct, err error)
	GetPrices() (prices ProductPrices, err error)
	GetPath() string
}

var ErrMissingCountryCode = errors.New("invalid code")

//
type PICSExtended map[string]string
type PICSAppCommon map[string]string
type PICSAppUFS map[string]string
type PICSController map[string]bool
type PICSAppConfig map[string]string
type PICSAppConfigLaunchItem struct {
	Order             int      `json:"order"`
	Executable        string   `json:"executable"`
	Arguments         string   `json:"arguments"`
	Description       string   `json:"description"`
	Typex             string   `json:"type"`
	OSList            string   `json:"oslist"`
	OSArch            string   `json:"osarch"`
	OwnsDLCs          []string `json:"ownsdlc"`
	BetaKey           string   `json:"betakey"`
	WorkingDir        string   `json:"workingdir"`
	VRMode            string   `json:"vrmode"`
	VACModuleFilename string   `json:"vacmodulefilename"`
}
type PICSDepots struct {
	Depots   []PICSAppDepotItem
	Branches []PICSAppDepotBranches
	Extra    map[string]string
}
type PICSAppDepotItem struct {
	ID                         int               `json:"id"`
	Name                       string            `json:"name"`
	Configs                    map[string]string `json:"config"`
	Manifests                  map[string]string `json:"manifests"`
	EncryptedManifests         string            `json:"encryptedmanifests"`
	MaxSize                    int64             `json:"maxsize"`
	App                        int               `json:"depotfromapp"`
	DLCApp                     int               `json:"dlcappid"`
	SystemDefined              bool              `json:"systemdefined"`
	Optional                   bool              `json:"optional"`
	SharedInstall              bool              `json:"sharedinstall"`
	SharedDepotType            bool              `json:"shareddepottype"`
	LVCache                    bool              `json:"lvcache"`
	AllowAddRemoveWhileRunning bool              `json:"allowaddremovewhilerunning"`
}
type PICSAppDepotBranches struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	BuildID          int    `json:"buildid"`
	TimeUpdated      int64  `json:"timeupdated"`
	PasswordRequired bool   `json:"pwdrequired"`
	LCSRequired      bool   `json:"lcsrequired"`
	DefaultForSubs   string `json:"defaultforsubs"`
	UnlockForSubs    string `json:"unlockforsubs"`
}

//
type ProductType string

const (
	ProductTypeApp     ProductType = "product"
	ProductTypePackage ProductType = "package"
)

//
type ProductPrices map[steam.CountryCode]ProductPriceStruct

func (p *ProductPrices) AddPriceFromPackage(code steam.CountryCode, prices steam.PackageDetailsBody) {

	if prices.Data.Price.Currency == "" {

		locale, err := helpers.GetLocaleFromCountry(code)
		log.Err(err)

		prices.Data.Price.Currency = string(locale.CurrencyCode)
	}

	(*p)[code] = ProductPriceStruct{
		Currency:        prices.Data.Price.Currency,
		Initial:         prices.Data.Price.Initial,
		Final:           prices.Data.Price.Final,
		DiscountPercent: prices.Data.Price.DiscountPercent,
		Individual:      prices.Data.Price.Individual,
	}
}

func (p *ProductPrices) AddPriceFromApp(code steam.CountryCode, prices steam.AppDetailsBody) {

	if prices.Data.PriceOverview.Currency == "" {

		locale, err := helpers.GetLocaleFromCountry(code)
		log.Err(err)

		prices.Data.PriceOverview.Currency = string(locale.CurrencyCode)
	}

	(*p)[code] = ProductPriceStruct{
		Currency:        prices.Data.PriceOverview.Currency,
		Initial:         prices.Data.PriceOverview.Initial,
		Final:           prices.Data.PriceOverview.Final,
		DiscountPercent: prices.Data.PriceOverview.DiscountPercent,
	}
}

func (p ProductPrices) Get(code steam.CountryCode) (price ProductPriceStruct, err error) {
	if val, ok := p[code]; ok {
		return val, err
	}
	return price, ErrMissingCountryCode
}

// ProductPriceStruct
type ProductPriceStruct struct {
	Currency        string `json:"currency"`
	Initial         int    `json:"initial"`
	Final           int    `json:"final"`
	DiscountPercent int    `json:"discount_percent"`
	Individual      int    `json:"individual"`
}

func (p ProductPriceStruct) GetInitial() string {

	code, err := helpers.GetLocaleFromCurrency(steam.CurrencyCode(p.Currency))
	log.Err(err)

	locale, err := helpers.GetLocaleFromCountry(code.CountryCode)
	log.Err(err)

	return locale.Format(p.Initial)
}

func (p ProductPriceStruct) GetFinal() string {

	code, err := helpers.GetLocaleFromCurrency(steam.CurrencyCode(p.Currency))
	log.Err(err)

	locale, err := helpers.GetLocaleFromCountry(code.CountryCode)
	log.Err(err)

	return locale.Format(p.Final)
}

func (p ProductPriceStruct) GetDiscountPercent() string {
	return strconv.Itoa(p.DiscountPercent) + "%"
}

func (p ProductPriceStruct) GetIndividual() string {

	code, err := helpers.GetLocaleFromCurrency(steam.CurrencyCode(p.Currency))
	log.Err(err)

	locale, err := helpers.GetLocaleFromCountry(code.CountryCode)
	log.Err(err)

	return locale.Format(p.Individual)
}

func (p ProductPriceStruct) GetCountryName(code steam.CountryCode) string {
	locale, err := helpers.GetLocaleFromCountry(code)
	log.Err(err)
	return locale.CountryName
}

func (p ProductPriceStruct) GetFlag(code steam.CountryCode) string {
	return "/assets/img/flags/" + strings.ToLower(string(code)) + ".png"
}

//
type ProductPriceFormattedStruct struct {
	Initial         string `json:"initial"`
	Final           string `json:"final"`
	DiscountPercent string `json:"discount_percent"`
	Individual      string `json:"individual"`
}

//
func GetPriceFormatted(product ProductInterface, code steam.CountryCode) (ret ProductPriceFormattedStruct) {

	price, err := product.GetPrice(code)
	if err == nil {

		locale, err := helpers.GetLocaleFromCountry(code)
		if err == nil {
			ret = ProductPriceFormattedStruct{
				Initial:         locale.Format(price.Initial),
				Final:           locale.Format(price.Final),
				DiscountPercent: locale.Format(price.DiscountPercent),
				Individual:      locale.Format(price.Individual),
			}
		}
	}

	return ret
}
