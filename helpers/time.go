package helpers

import (
	"time"

	"github.com/Jleagle/go-durationfmt"
	"github.com/gamedb/website/log"
)

const (
	Date         = "02 Jan"
	DateTime     = "02 Jan 15:04"
	DateYear     = "02 Jan 06"
	DateYearTime = "02 Jan 06 15:04"
)

func GetTimeShort(minutes int, max int) (ret string) {

	t, err := durationfmt.Format(time.Minute*time.Duration(minutes), "%d days, %h hours")
	log.Err(err)

	return t
}

func GetTimeLong(minutes int, max int) (ret string) {

	t, err := durationfmt.Format(time.Minute*time.Duration(minutes), "%d, %hh")
	log.Err(err)

	return t
}
