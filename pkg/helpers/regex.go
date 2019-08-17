package helpers

import (
	"regexp"
)

var (
	RegexIntsOnly   = regexp.MustCompile("[^0-9]+")
	RegexTimestamps = regexp.MustCompile("1[0-9]]{9}")
)
