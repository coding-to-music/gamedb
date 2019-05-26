package helpers

import (
	"bytes"
	"math/rand"
	"regexp"
	"strings"
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

func JoinInterface(i []interface{}) string {

	var stringSlice []string

	for _, v := range i {

		s, ok := v.(string)
		if ok {
			stringSlice = append(stringSlice, s)
		}
	}

	return strings.Join(stringSlice, " | ")
}

const Letters = "abcdefghijklmnopqrstuvwxyz"
const Numbers = "0123456789"

func RandString(n int, chars string) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

func InsertNewLines(s string, n int) string {
	var buffer bytes.Buffer
	var n1 = n - 1
	var l1 = len(s) - 1
	for i, r := range s {
		buffer.WriteRune(r)
		if i%n == n1 && i != l1 {
			buffer.WriteString("<wbr />")
		}
	}
	return buffer.String()
}
