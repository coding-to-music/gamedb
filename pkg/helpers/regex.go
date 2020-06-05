package helpers

import (
	"regexp"
)

var (
	RegexInts                  = regexp.MustCompile("^[0-9]+$")
	RegexNonInts               = regexp.MustCompile("[^0-9]+")
	RegexNonNumbers            = regexp.MustCompile("[^0-9-]+")
	RegexTimestamps            = regexp.MustCompile("1[0-9]{9}")
	RegexNonAlphaNumeric       = regexp.MustCompile("[^a-zA-Z0-9]+")
	RegexNonAlphaNumericSpace  = regexp.MustCompile("[^a-zA-Z0-9 ]+")
	RegexMultipleNewLines      = regexp.MustCompile("[\n]{3,}")
	RegexNewLine               = regexp.MustCompile("[\n\r]")
	RegexFilterEmptyCharacters = regexp.MustCompile(`[\p{Cf}\p{Co}\p{Cs}\p{Cc}\p{C}]`)
)
