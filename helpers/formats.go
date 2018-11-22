package helpers

import (
	"strconv"
)

// todo, these are all not needed

func CentsInt(cents int) float64 {
	return DollarsFloat(float64(cents) / 100)
}

func DollarsFloat(dollars float64) float64 {

	float, err := strconv.ParseFloat(IntToFloat(dollars, 2), 64)
	if err != nil {
		return 0
	}

	return float
}
