package sql

import (
	"math"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/gamedb/gamedb/pkg/helpers"
)

func GetMeanPrice(code steamapi.ProductCC, prices string) (string, error) {

	means := map[steamapi.ProductCC]float64{}

	err := helpers.Unmarshal([]byte(prices), &means)
	if err == nil {
		if val, ok := means[code]; ok {
			return helpers.FormatPrice(helpers.GetProdCC(code).CurrencyCode, int(math.Round(val))), err
		}
	}

	return "-", err
}
