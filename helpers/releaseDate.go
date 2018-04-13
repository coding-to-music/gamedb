package helpers

import (
	"errors"
	"time"
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

func GetReleaseDateNice(date string) (out string) {

	t, err := getReleaseDate(date)
	if err != nil {
		return date
	}

	return t.Format(DateYear)
}

func GetReleaseDateUnix(date string) int64 {

	t, err := getReleaseDate(date)
	if err != nil {
		return 0
	}

	return t.Unix()
}
