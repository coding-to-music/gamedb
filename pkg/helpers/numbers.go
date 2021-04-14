package helpers

import (
	"math"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/log"
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

func GetTrendValue(i float64) string {

	i = RoundFloatTo2DP(i * 1000)

	if i > 0 {
		return "+" + humanize.Commaf(i)
	} else if i == 0 {
		return "-"
	} else {
		return humanize.Commaf(i)
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

func ChunkInt64s(ints []int64, n int) (chunks [][]int64) {

	for i := 0; i < len(ints); i += n {
		end := i + n

		if end > len(ints) {
			end = len(ints)
		}

		chunks = append(chunks, ints[i:end])
	}
	return chunks
}

func StringToInt(s string) int {

	s = RegexNonInts.ReplaceAllString(s, "")

	i, err := strconv.Atoi(s)
	if err != nil {
		log.ErrS(err)
	}

	return i
}

func Min(vars ...float64) float64 {

	if len(vars) == 0 {
		return 0
	}

	min := vars[0]
	for _, i := range vars {
		if min > i {
			min = i
		}
	}
	return min
}

func Max(vars ...float64) float64 {

	if len(vars) == 0 {
		return 0
	}

	max := vars[0]
	for _, i := range vars {
		if max < i {
			max = i
		}
	}
	return max
}

func Avg(vars ...float64) float64 {

	if len(vars) == 0 {
		return 0
	}

	var count float64
	var sum float64

	for _, i := range vars {
		count += 1
		sum += i
	}
	return sum / count
}
