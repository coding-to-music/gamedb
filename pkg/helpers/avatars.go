package helpers

import (
	"math"
)

// Steam's generated avatar
func GetAvatar2(level int) string {

	ret := "avatar2"

	n100 := math.Floor(float64(level)/100) * 100
	if n100 >= 100 {

		ret += " lvl_" + FloatToString(n100, 0)

		n10 := math.Floor(float64(level)/10) * 10
		n10String := FloatToString(n10, 0)
		n10String = n10String[len(n10String)-2:]

		if n10String != "00" {
			ret += " lvl_plus_" + n10String
		}
	}

	return ret
}
