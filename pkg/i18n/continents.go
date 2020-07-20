package i18n

import (
	"github.com/pariz/gountries"
)

// Continents
const (
	ContinentAfrica       = "AF"
	ContinentAntarctica   = "AN"
	ContinentAsia         = "AS"
	ContinentEurope       = "EU"
	ContinentNorthAmerica = "NA"
	ContinentSouthAmerica = "SA"
	ContinentOceania      = "OC"
)

type Continent struct {
	Key   string `json:"k"`
	Value string `json:"v"`
}

// These strings must match the continents in the gountries library
var Continents = []Continent{
	{Key: ContinentAfrica, Value: "Africa"},
	{Key: ContinentAntarctica, Value: "Antarctica"},
	{Key: ContinentAsia, Value: "Asia"},
	{Key: ContinentEurope, Value: "Europe"},
	{Key: ContinentNorthAmerica, Value: "North America"},
	{Key: ContinentSouthAmerica, Value: "South America"},
	// {Key: ContinentOceania, Value: "Oceania"},
}

//
func CountriesInContinent(continent string) (ret []string) {

	countries := gountriesInstance.FindCountries(gountries.Country{Geo: gountries.Geo{Continent: continent}})
	for _, country := range countries {
		ret = append(ret, country.Alpha2)
	}
	return ret
}
