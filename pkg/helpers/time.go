package helpers

import (
	"regexp"
	"strings"
	"time"

	"github.com/Jleagle/go-durationfmt"
	"github.com/gamedb/gamedb/pkg/log"
)

const (
	Date         = "02 Jan"
	DateTime     = "02 Jan 15:04"
	DateYear     = "02 Jan 2006"
	DateYearTime = "02 Jan 06 15:04"
	DateSQL      = "2006-01-02 15:04:05"
	DateSQLDay   = "2006-01-02"
)

func GetTimeShort(minutes int, max int) (ret string) {
	return formatTime(minutes, max, "%yy %om %ww %dd %hh %mm")
}

func GetTimeLong(minutes int, max int) (ret string) {
	return formatTime(minutes, max, "%y years %o months %w weeks %d days %h hours %m minutes")
}

var getTimeRegex = regexp.MustCompile(`([0-9]+\s?[a-z]+)`)

func formatTime(minutes int, pieces int, format string) string {

	if minutes == 0 {
		return "-"
	}

	t, err := durationfmt.Format(time.Minute*time.Duration(minutes), format)
	if err != nil {
		log.Err(err)
		return "-"
	}

	var ret []string
	for _, v := range getTimeRegex.FindAllString(t, -1) {

		if pieces > 0 && len(ret) >= pieces {
			break
		}

		if string(v[0]) == "0" {
			continue
		}

		if RegexNonInts.ReplaceAllString(v, "") == "1" {
			v = strings.TrimSuffix(v, "s")
		}

		ret = append(ret, v)
	}

	return strings.Join(ret, ", ")
}
