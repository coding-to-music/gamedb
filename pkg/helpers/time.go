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
	DateYear     = "02 Jan 06"
	DateYearTime = "02 Jan 06 15:04"
)

func GetTimeShort(minutes int, max int) (ret string) {
	return formatTime(minutes, max, "%yy %om %ww %dd %hh %mm", false)
}

func GetTimeLong(minutes int, max int) (ret string) {
	return formatTime(minutes, max, "%yyears %omonths %wweeks %ddays %hhours %mminutes", true)
}

func formatTime(minutes int, pieces int, format string, addSpaces bool) string {

	if minutes == 0 {
		return "-"
	}

	t, err := durationfmt.Format(time.Minute*time.Duration(minutes), format)
	log.Err(err)

	// Remove pieces that are zero-value
	var re = regexp.MustCompile(`(^0[a-z]+)|( 0[a-z]+)`)
	t = re.ReplaceAllString(t, "")

	//
	var re2 = regexp.MustCompile(`([0-9]+)([a-z]+?)(s?)$`)
	var ts = strings.Split(t, " ")

	var ret []string
	for _, v := range ts {

		if pieces > 0 && len(ret) >= pieces {
			break
		}

		v = strings.TrimSpace(v)
		if v != "" {

			// Add/remove plural "s"
			var s = "$3"
			if re2.ReplaceAllString(v, "$1") == "1" {
				s = ""
			}

			// Add/remove space
			if addSpaces {
				v = re2.ReplaceAllString(v, "$1 $2"+s)
			} else {
				v = re2.ReplaceAllString(v, "$1$2"+s)
			}

			ret = append(ret, v)
		}
	}

	return strings.Join(ret, ", ")
}
