package helpers

import (
	"fmt"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/logging"
	"github.com/pariz/gountries"
	"golang.org/x/text/currency"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var gountriesInstance = gountries.New()
var byCurrency = map[steam.CurrencyCode]Locale{}
var byCountry = map[steam.CountryCode]Locale{}
var locales = []Locale{
	{CountryCode: steam.CountryAE, CurrencyCode: steam.CurrencyAED, CurrencySymbol: "D"},
	{CountryCode: steam.CountryAR, CurrencyCode: steam.CurrencyARS, CurrencySymbol: "$"},
	{CountryCode: steam.CountryAU, CurrencyCode: steam.CurrencyAUD, CurrencySymbol: "$"},
	{CountryCode: steam.CountryBR, CurrencyCode: steam.CurrencyBRL, CurrencySymbol: "$"},
	{CountryCode: steam.CountryCA, CurrencyCode: steam.CurrencyCAD, CurrencySymbol: "$"},
	{CountryCode: steam.CountryCH, CurrencyCode: steam.CurrencyCHF, CurrencySymbol: "Fr"},
	{CountryCode: steam.CountryCL, CurrencyCode: steam.CurrencyCLP, CurrencySymbol: "$"},
	{CountryCode: steam.CountryCN, CurrencyCode: steam.CurrencyCNY, CurrencySymbol: "¥"},
	{CountryCode: steam.CountryCO, CurrencyCode: steam.CurrencyCOP, CurrencySymbol: "$"},
	{CountryCode: steam.CountryCR, CurrencyCode: steam.CurrencyCRC, CurrencySymbol: "₡"},
	{CountryCode: steam.CountryDE, CurrencyCode: steam.CurrencyEUR, CurrencySymbol: "€"},
	{CountryCode: steam.CountryGB, CurrencyCode: steam.CurrencyGBP, CurrencySymbol: "£"},
	{CountryCode: steam.CountryHK, CurrencyCode: steam.CurrencyHKD, CurrencySymbol: "$"},
	{CountryCode: steam.CountryIL, CurrencyCode: steam.CurrencyILS, CurrencySymbol: "₪"},
	{CountryCode: steam.CountryID, CurrencyCode: steam.CurrencyIDR, CurrencySymbol: "Rp"},
	{CountryCode: steam.CountryIN, CurrencyCode: steam.CurrencyINR, CurrencySymbol: "₹"},
	{CountryCode: steam.CountryJP, CurrencyCode: steam.CurrencyJPY, CurrencySymbol: "¥"},
	{CountryCode: steam.CountryKR, CurrencyCode: steam.CurrencyKRW, CurrencySymbol: "₩"},
	{CountryCode: steam.CountryKW, CurrencyCode: steam.CurrencyKWD, CurrencySymbol: "KD"},
	{CountryCode: steam.CountryKZ, CurrencyCode: steam.CurrencyKZT, CurrencySymbol: "₸"},
	{CountryCode: steam.CountryMX, CurrencyCode: steam.CurrencyMXN, CurrencySymbol: "$"},
	{CountryCode: steam.CountryMY, CurrencyCode: steam.CurrencyMYR, CurrencySymbol: "RM"},
	{CountryCode: steam.CountryNO, CurrencyCode: steam.CurrencyNOK, CurrencySymbol: "kr"},
	{CountryCode: steam.CountryNZ, CurrencyCode: steam.CurrencyNZD, CurrencySymbol: "$"},
	{CountryCode: steam.CountryPE, CurrencyCode: steam.CurrencyPEN, CurrencySymbol: "S"},
	{CountryCode: steam.CountryPH, CurrencyCode: steam.CurrencyPHP, CurrencySymbol: "₱"},
	{CountryCode: steam.CountryPL, CurrencyCode: steam.CurrencyPLN, CurrencySymbol: "zł"},
	{CountryCode: steam.CountryQA, CurrencyCode: steam.CurrencyQAR, CurrencySymbol: "QR"},
	{CountryCode: steam.CountryRU, CurrencyCode: steam.CurrencyRUB, CurrencySymbol: "₽"},
	{CountryCode: steam.CountrySA, CurrencyCode: steam.CurrencySAR, CurrencySymbol: "SR"},
	{CountryCode: steam.CountrySG, CurrencyCode: steam.CurrencySGD, CurrencySymbol: "$"},
	{CountryCode: steam.CountryTH, CurrencyCode: steam.CurrencyTHB, CurrencySymbol: "฿"},
	{CountryCode: steam.CountryTR, CurrencyCode: steam.CurrencyTRY, CurrencySymbol: "₺"},
	{CountryCode: steam.CountryTW, CurrencyCode: steam.CurrencyTWD, CurrencySymbol: "$"},
	{CountryCode: steam.CountryUA, CurrencyCode: steam.CurrencyUAH, CurrencySymbol: "₴"},
	{CountryCode: steam.CountryUS, CurrencyCode: steam.CurrencyUSD, CurrencySymbol: "$"},
	{CountryCode: steam.CountryUY, CurrencyCode: steam.CurrencyUYU, CurrencySymbol: "$"},
	{CountryCode: steam.CountryVN, CurrencyCode: steam.CurrencyVND, CurrencySymbol: "₫"},
	{CountryCode: steam.CountryZA, CurrencyCode: steam.CurrencyZAR, CurrencySymbol: "R"},
}

func init() {
	for _, v := range locales {
		byCurrency[v.CurrencyCode] = v
		byCountry[v.CountryCode] = v
	}
}

func GetFromCurrency(code steam.CurrencyCode) (loc Locale, err error) {

	if val, ok := byCurrency[code]; ok {
		return val, err
	}
	return loc, err
}

func GetFromCountry(code steam.CountryCode) (loc Locale, err error) {

	if val, ok := byCountry[code]; ok {
		return val, err
	}
	return loc, err
}

type Locale struct {
	CountryCode    steam.CountryCode
	CurrencyCode   steam.CurrencyCode
	CurrencySymbol string
	CountryID      steam.Language
}

func (l Locale) GetCountryName() string {

	if l.CountryCode == "" {
		return string(l.CountryCode)
	} else if l.CountryCode == "BQ" {
		return "Bonaire, Sint Eustatius and Saba"
	}

	country, err := gountriesInstance.FindCountryByAlpha(string(l.CountryCode))
	if err != nil {
		logging.Error(err)
		return string(l.CountryCode)
	}

	return country.Name.Common
}

func (l Locale) Format(amount int, lang ...string) (ret string, err error) {

	unit, err := currency.ParseISO(string(l.CurrencyCode))
	if err != nil {
		return ret, err
	}
	a := unit.Amount(float64(amount) / 100)
	symbol := currency.NarrowSymbol(a)

	// Format as if you were this language ("en-AU")
	if len(lang) > 0 {
		tag, err := language.Parse(lang[0])
		if err != nil {
			return ret, err
		}
		printer := message.NewPrinter(tag)
		return printer.Sprint(symbol), err
	}

	return fmt.Sprint(symbol), err
}
