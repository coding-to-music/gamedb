package i18n

import (
	"errors"

	"github.com/gamedb/gamedb/pkg/log"
	"go.uber.org/zap"
)

func CountryCodeToName(code string) string {

	switch code {
	case "", "_", "0":
		return "No Country"
	case "AN":
		return "Netherlands Antilles"
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
	case "TP":
		return "East Timor"
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
	case "", "0":
		return "", nil
	case "BQ":
		return ContinentSouthAmerica, nil
	case "SH":
		return ContinentAfrica, nil
	case "TP":
		return ContinentAsia, nil
	case "YU", "FX", "XK", "AN":
		return ContinentEurope, nil
	}

	country, err := gountriesInstance.FindCountryByAlpha(code)
	if err != nil {
		return "", err
	}

	for _, v := range Continents {
		if v.Value == country.Geo.Continent {
			return v.Key, nil
		}
	}

	return "", errors.New("unknown continent")
}
