package helpers

import (
	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/logging"
	"github.com/leekchan/accounting"
	"github.com/pariz/gountries"
)

// https://partner.steamgames.com/doc/store/localization

var Languages = map[string]Language{}

type Language struct {
	EnglishName string
	Native      string
	Code        string
}

func CountryCodeToName(code string) string {

	if code == "" {
		return code
	} else if code == "BQ" {
		return "Bonaire, Sint Eustatius and Saba"
	}

	query := gountries.New()
	country, err := query.FindCountryByAlpha(code)
	if err != nil {
		logging.Error(err)
		return code
	}

	return country.Name.Common
}

func CurrencyFormat(code steam.CountryCode, amount int) string {

	ac := accounting.Accounting{Symbol: CurrencySymbol(code), Precision: 2}
	return ac.FormatMoney(float64(amount) / 100)
}

func CurrencySymbol(code steam.CountryCode) string {

	var x = map[steam.CountryCode]string{
		steam.CountryAE: "D",
		steam.CountryAR: "$",
		steam.CountryAU: "A$",
		steam.CountryBR: "R$",
		steam.CountryCA: "C$",
		steam.CountryCH: "Fr.",
		steam.CountryCL: "$",
		steam.CountryCN: "¥",
		steam.CountryCO: "$",
		steam.CountryCR: "₡",
		steam.CountryDE: "€",
		steam.CountryGB: "£",
		steam.CountryHK: "HK$",
		steam.CountryIL: "₪",
		steam.CountryID: "Rp",
		steam.CountryIN: "₹",
		steam.CountryJP: "¥",
		steam.CountryKR: "₩",
		steam.CountryKW: "KD",
		steam.CountryKZ: "₸",
		steam.CountryMX: "Mex$",
		steam.CountryMY: "RM",
		steam.CountryNO: "kr",
		steam.CountryNZ: "$",
		steam.CountryPE: "S/",
		steam.CountryPH: "₱",
		steam.CountryPL: "zł",
		steam.CountryQA: "QR",
		steam.CountryRU: "₽",
		steam.CountrySA: "SR",
		steam.CountrySG: "S$",
		steam.CountryTH: "฿",
		steam.CountryTR: "₺",
		steam.CountryTW: "NT$",
		steam.CountryUA: "₴",
		steam.CountryUS: "$",
		steam.CountryUY: "$U",
		steam.CountryVN: "₫",
		steam.CountryZA: "R",
	}

	if val, ok := x[code]; ok {
		return val
	}

	return "$"
}
