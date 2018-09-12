package helpers

import (
	"github.com/pariz/gountries"
	"github.com/steam-authority/steam-authority/logger"
)

// https://partner.steamgames.com/doc/store/localization

var Languages = map[string]Language{}

type Language struct {
	EnglishName string
	Native      string
	Code        string
}

func CountryCodeToName(code string) string {

	query := gountries.New()
	country, err := query.FindCountryByAlpha(code)
	if err != nil {
		logger.Error(err)
		return code
	}

	return country.Name.Common
}
