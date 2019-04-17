package helpers

import (
	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/pkg"
)

func GetMeanPrice(code steam.CountryCode, prices string) (string, error) {

	means := map[steam.CountryCode]float64{}

	locale, err := pkg.GetLocaleFromCountry(code)
	log.Err(err)

	err = helpers.Unmarshal([]byte(prices), &means)
	if err == nil {
		if val, ok := means[code]; ok {
			return locale.CurrencySymbol + helpers.FloatToString(helpers.RoundFloatTo2DP(float64(val)/100), 2), err
		}
	}

	return locale.CurrencySymbol + "0", err
}
