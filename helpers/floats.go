package helpers

import (
	"fmt"
	"strconv"
)

func IntToFloat(float float64, after int) string {
	return fmt.Sprintf("%0."+strconv.Itoa(after)+"f", float)
}

func PadInt(i, padding int) string {
	return fmt.Sprintf("%"+strconv.Itoa(padding)+"d", i)
}
