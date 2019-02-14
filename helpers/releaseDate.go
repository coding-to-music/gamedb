package helpers

import (
	"errors"
	"time"

	"github.com/jinzhu/now"
)

func getReleaseDate(date string) (t time.Time, err error) {

	if date == "" {
		return t, errors.New("blank")
	}

	t, err = time.Parse("Jan 2, 2006", date)
	if err != nil {

		t, err = time.Parse("2 Jan, 2006", date)
		if err != nil {

			t, err = time.Parse("Jan 2006", date)
			if err != nil {

				t, err = time.Parse("2006", date)
			}
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

	days := release.Sub(now.BeginningOfDay()).Hours() / 24

	if days == 0 {
		return "Today"
	}

	return "In " + GetTimeLong(int(release.Sub(now.BeginningOfDay()).Minutes()), 2)
}
