package helpers

import (
	"regexp"
)

func TruncateString(str string, size int) string {
	ret := str
	if len(str) > size {
		if size > 3 {
			size -= 3
		}
		ret = str[0:size] + "..."
	}
	return ret
}

func GetHashTag(string string) (ret string) {

	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		return ""
	}

	return "#" + reg.ReplaceAllString(string, "")
}
