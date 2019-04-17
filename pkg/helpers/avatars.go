package helpers

import (
	"math"

	"github.com/gamedb/website/pkg"
)

// Steam's generated avatar
func GetAvatar2(level int) string {

	ret := "avatar2"

	n100 := math.Floor(float64(level)/100) * 100
	if n100 >= 100 {

		ret += " lvl_" + helpers.FloatToString(n100, 0)

		n10 := math.Floor(float64(level)/10) * 10
		n10String := helpers.FloatToString(n10, 0)
		n10String = n10String[len(n10String)-2:]

		if n10String != "00" {
			ret += " lvl_plus_" + n10String
		}
	}

	return ret
}
