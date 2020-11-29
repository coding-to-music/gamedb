package helpers

import (
	"regexp"
)

//noinspection GoUnusedGlobalVariable
var (
	RegexInts                  = regexp.MustCompile(`[0-9]+`)
	RegexIntsOnly              = regexp.MustCompile(`^[0-9]+$`)
	RegexMD5Only               = regexp.MustCompile(`^[a-f0-9]{32}$`)
	RegexSha1                  = regexp.MustCompile(`[a-f0-9]{40}`)
	RegexSha1Only              = regexp.MustCompile(`^[a-f0-9]{40}$`)
	RegexTimestamps            = regexp.MustCompile(`1[0-9]{9}`)
	RegexNonInts               = regexp.MustCompile(`[^0-9]+`)
	RegexNonNumbers            = regexp.MustCompile(`[^0-9-]+`)
	RegexNonAlphaNumeric       = regexp.MustCompile(`[^a-zA-Z0-9]+`)
	RegexNonAlphaNumericSpace  = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)
	RegexNewLines              = regexp.MustCompile(`[\n\r]{3,}`)
	RegexNewLine               = regexp.MustCompile(`[\n\r]`)
	RegexSpaces                = regexp.MustCompile(`[\s]`)
	RegexSpacesStartEnd        = regexp.MustCompile(`^[\s\n\r]+|[\s\n\r]+$`)
	RegexFilterEmptyCharacters = regexp.MustCompile(`[\p{Cf}\p{Co}\p{Cs}\p{Cc}\p{C}\p{Braille}]`)
	RegexSmallRomanOnly        = regexp.MustCompile(`^[IVX]+$`)
	RegexYouTubeID             = regexp.MustCompile(`[a-zA-Z0-9_\-]{11}`)
	RegexIP                    = regexp.MustCompile(`(?:[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3})`)
)
