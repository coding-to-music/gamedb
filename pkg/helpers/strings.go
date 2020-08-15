package helpers

import (
	"math/rand"
	"strconv"
	"strings"
)

func TruncateString(str string, size int, tail string) string {
	ret := str
	if len(str) > size {
		if size > len(tail) {
			size -= len(tail)
		}
		ret = strings.TrimSpace(str[0:size]) + tail
	}
	return ret
}

func GetHashTag(string string) (ret string) {
	return "#" + RegexNonAlphaNumeric.ReplaceAllString(string, "")
}

func JoinInts(i []int, sep string) string {

	var stringSlice []string

	for _, v := range i {
		stringSlice = append(stringSlice, strconv.Itoa(v))
	}

	return strings.Join(stringSlice, sep)
}

func JoinInt64s(i []int64, sep string) string {

	var stringSlice []string

	for _, v := range i {
		stringSlice = append(stringSlice, strconv.FormatInt(v, 10))
	}

	return strings.Join(stringSlice, sep)
}

const Letters = "abcdefghijklmnopqrstuvwxyz"
const LettersCaps = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const Numbers = "0123456789"

func RandString(n int, chars string) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
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
