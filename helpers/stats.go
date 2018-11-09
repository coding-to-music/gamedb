package helpers

import (
	"fmt"

	"github.com/Jleagle/steam-go/steam"
)

func GetMeanPrice(code steam.CountryCode, json string) (string, error) {

	means := map[steam.CountryCode]int{}

	symbol := CurrencySymbol(code)

	err := Unmarshal([]byte(json), &means)
	if err == nil {
		if val, ok := means[code]; ok {
			return symbol + fmt.Sprintf("%0.2f", float64(val)/100), err
		}
	}

	return symbol + "0", err
}

func GetMeanScore(code steam.CountryCode, json string) (string, error) {

	means := map[steam.CountryCode]float64{}

	err := Unmarshal([]byte(json), &means)
	if err == nil {
		if val, ok := means[code]; ok {
			return fmt.Sprintf("%0.2f", val) + "%", err
		}
	}

	return "0%", err
}
