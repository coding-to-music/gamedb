package helpers

import (
	"github.com/pariz/gountries"
	"github.com/steam-authority/steam-authority/logging"
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
