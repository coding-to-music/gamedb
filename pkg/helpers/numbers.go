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

// keeps extra zeros
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

func ShortHandNumber(i int64) string {

	if i >= 1000000000 {
		return humanize.FormatFloat("#,###.###", float64(i)/1000000000) + "B"
	}

	if i >= 1000000 {
		return humanize.FormatFloat("#,###.##", float64(i)/1000000) + "M"
	}

	if i >= 1000 {
		return humanize.FormatFloat("#,###.#", float64(i)/1000) + "K"
	}

	return humanize.FormatFloat("", float64(i))
}

func RoundTo(x, unit float64) float64 {
	return math.Round(x/unit) * unit
}

func ChunkInts(ints []int, n int) (chunks [][]int) {

	for i := 0; i < len(ints); i += n {
		end := i + n

		if end > len(ints) {
			end = len(ints)
		}

		chunks = append(chunks, ints[i:end])
	}
	return chunks
}
