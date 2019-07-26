package helpers

import (
	"errors"
	"math"
	"time"

	"github.com/jinzhu/now"
)

var releaseDates = []string{
	"Jan 2, 2006",
	"2 Jan, 2006",
	"Jan 2006",
	"2006",
}

func getReleaseDate(date string) (t time.Time, err error) {

	if date == "" {
		return t, errors.New("blank")
	}

	for _, v := range releaseDates {
		t, err = time.Parse(v, date)
		if err == nil {
			break
		}
	}

	return t, err
}

func GetReleaseDateUnix(date string) int64 {

	t, err := getReleaseDate(date)
	if err != nil {
		return 0
	}

	return t.Unix()
}

func GetDaysToRelease(unix int64) string {

	release := time.Unix(unix, 0)

	days := math.Floor(release.Sub(now.BeginningOfDay()).Hours() / 24)

	if days == 0 {
		return "Today"
	}

	return "In " + GetTimeLong(int(days)*24*60, 2)
}
