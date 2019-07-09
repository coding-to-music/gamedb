package helpers

import (
	"math/rand"
	"regexp"
	"strconv"
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

func JoinInts(i []int) string {

	var stringSlice []string

	for _, v := range i {
		stringSlice = append(stringSlice, strconv.Itoa(v))
	}

	return strings.Join(stringSlice, ",")
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

var newLineRegex = regexp.MustCompile("(.{5})")

func InsertNewLines(s string) string {
	return newLineRegex.ReplaceAllString(s, "$1<wbr />")
}

func ChunkStrings(strings []string, n int) (chunks [][]string) {

	for i := 0; i < len(strings); i += n {
		end := i + n

		if end > len(strings) {
			end = len(strings)
		}

		chunks = append(chunks, strings[i:end])
	}
	return chunks
}
