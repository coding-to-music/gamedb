package i18n

import (
	"errors"

	"github.com/gamedb/gamedb/pkg/log"
	"go.uber.org/zap"
)

func CountryCodeToName(code string) string {

	switch code {
	case "", "_":
		return "No Country"
	case "AX":
		return "Aland Islands"
	case "BQ":
		return "Bonaire, Sint Eustatius and Saba"
	case "SH":
		return "Saint Helena"
	case "XK":
		return "Kosovo"
	case "FX":
		return "France, Metropolitan"
	case "YU":
		return "Yugoslavia"
	}

	country, err := gountriesInstance.FindCountryByAlpha(code)
	if err != nil {
		log.Err(err.Error(), zap.String("code", code))
		return code
	}

	return country.Name.Common
}

func CountryCodeToContinent(code string) (string, error) {

	switch code {
	case "":
		return "", nil
	case "BQ":
		return ContinentSouthAmerica, nil
	case "SH":
		return ContinentAfrica, nil
	case "YU", "FX", "XK":
		return ContinentEurope, nil
	}

	country, err := gountriesInstance.FindCountryByAlpha(code)
	if err != nil {
		return "", err
	}

	for _, v := range Continents {
		if v.Value == country.Continent {
			return v.Key, nil
		}
	}

	return "", errors.New("unknown continent")
}
