package helpers

import (
	"fmt"
	"strconv"
)

func CentsInt(cents int) float64 {
	return DollarsFloat(float64(cents) / 100)
}

func CentsFloat(cents float64) float64 {
	return DollarsFloat(cents / 100)
}

func DollarsFloat(dollars float64) float64 {

	float, err := strconv.ParseFloat(Round2(dollars), 64)
	if err != nil {
		return 0
	}

	return float
}

func Round2(float float64) string {
	return fmt.Sprintf("%0.2f", float)
}
