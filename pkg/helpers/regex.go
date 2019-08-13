package helpers

import (
	"regexp"
)

var (
	RegexIntsOnly = regexp.MustCompile("[^0-9]+")
)
