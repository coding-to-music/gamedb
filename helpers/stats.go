package helpers

import (
	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/log"
)

func GetMeanPrice(code steam.CountryCode, prices string) (string, error) {

	means := map[steam.CountryCode]float64{}

	locale, err := GetLocaleFromCountry(code)
	log.Log(err)

	err = Unmarshal([]byte(prices), &means)
	if err == nil {
		if val, ok := means[code]; ok {
			return locale.CurrencySymbol + FloatToString(RoundFloatTo2DP(float64(val)/100), 2), err
		}
	}

	return locale.CurrencySymbol + "0", err
}

func GetMeanScore(code steam.CountryCode, scores string) (string, error) {

	means := map[steam.CountryCode]float64{}

	err := Unmarshal([]byte(scores), &means)
	if err == nil {
		if val, ok := means[code]; ok {
			return FloatToString(RoundFloatTo2DP(val), 2) + "%", err
		}
	}

	return "0%", err
}
