package helpers

import (
	"math"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
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

func OrdinalComma(i int) string {

	iString := strconv.Itoa(i)

	ord := humanize.Ordinal(i)
	ord = strings.Replace(ord, iString, "", 1)

	return humanize.Comma(int64(i)) + ord
}

func TrendValue(i int64) string {

	if i > 0 {
		return "+" + humanize.Comma(i)
	} else if i == 0 {
		return "-"
	} else {
		return humanize.Comma(i)
	}
}

func PercentageChange(old, new int) (delta float64) {
	diff := float64(new - old)
	delta = (diff / float64(old)) * 100
	return
}
