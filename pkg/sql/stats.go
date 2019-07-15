package sql

import (
	"math"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/gamedb/pkg/helpers"
)

func GetMeanPrice(code steam.ProductCC, prices string) (string, error) {

	means := map[steam.ProductCC]float64{}

	err := helpers.Unmarshal([]byte(prices), &means)
	if err == nil {
		if val, ok := means[code]; ok {
			return helpers.FormatPrice(helpers.ProdCCToStruct(code).CurrencyCode, int(math.Round(val))), err
		}
	}

	return "-", err
}
