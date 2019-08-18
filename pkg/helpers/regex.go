package helpers

import (
	"regexp"
)

var (
	RegexNonInts    = regexp.MustCompile("[^0-9]+")
	RegexNonNumbers = regexp.MustCompile("[^0-9-]+")
	RegexTimestamps = regexp.MustCompile("1[0-9]{9}")
)
