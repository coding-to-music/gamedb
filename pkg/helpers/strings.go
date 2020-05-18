package helpers

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/gamedb/gamedb/pkg/log"
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

func InterfaceToString(i interface{}) string {
	switch i.(type) {
	case time.Duration:
		return i.(time.Duration).String()
	case time.Time:
		return i.(time.Time).String()
	case bool:
		return strconv.FormatBool(i.(bool))
	case int:
		return strconv.Itoa(i.(int))
	case int64:
		return strconv.FormatInt(i.(int64), 10)
	case string:
		return i.(string)
	case []interface{}:
		var sli []string
		for _, v := range i.([]interface{}) {
			sli = append(sli, InterfaceToString(v))
		}
		return "(" + strings.Join(sli, ",") + ")"
	default:
		log.Info("Can't convert val to " + fmt.Sprintf("%T", i))
		return ""
	}
}

func JoinInts(i []int, sep string) string {

	var stringSlice []string

	for _, v := range i {
		stringSlice = append(stringSlice, strconv.Itoa(v))
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
