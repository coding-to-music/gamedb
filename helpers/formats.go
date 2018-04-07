package helpers

import "fmt"

func CentsInt(cents int) (string) {
	return DollarsFloat(float64(cents) / 100)
}

func CentsFloat(cents float64) (string) {
	return DollarsFloat(cents / 100)
}

func DollarsFloat(dollars float64) (string) {
	return fmt.Sprintf("%0.2f", dollars)
}
