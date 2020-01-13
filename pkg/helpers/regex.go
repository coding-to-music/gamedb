package helpers

import (
	"regexp"
)

var (
	RegexNumbers              = regexp.MustCompile("^[0-9]+$")
	RegexNonInts              = regexp.MustCompile("[^0-9]+")
	RegexNonNumbers           = regexp.MustCompile("[^0-9-]+")
	RegexTimestamps           = regexp.MustCompile("1[0-9]{9}")
	RegexNonAlphaNumeric      = regexp.MustCompile("[^a-zA-Z0-9]+")
	RegexNonAlphaNumericSpace = regexp.MustCompile("[^a-zA-Z0-9 ]+")
)
