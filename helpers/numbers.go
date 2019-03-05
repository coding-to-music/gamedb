package helpers

import (
	"math"
	"strconv"
)

func RoundIntTo2DP(i int) float64 {
	return float64(i) / 100
}

func RoundFloatTo1DP(f float64) float64 {
	return math.Round(f*10) / 10
}

func RoundFloatTo2DP(f float64) float64 {
	return math.Round(f*100) / 100
}

func FloatToString(f float64, decimals int) string {
	return strconv.FormatFloat(f, 'f', decimals, 64)
}
