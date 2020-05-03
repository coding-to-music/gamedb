package helpers

import (
	"math"
	"time"

	"github.com/jinzhu/now"
)

var releaseDateFormats = []string{
	"2 Jan 2006",
	"2 Jan, 2006",
	"Jan 2, 2006",
	"Jan 2006",
	"January 2, 2006",
	"January 2006",
	"2006",
}

func GetReleaseDateUnix(date string) int64 {

	// for k, v := range map[string]string{"Q1 ": "January ", "Q2 ": "April ", "Q3 ": "July ", "Q4 ": "October "} {
	// 	if strings.HasPrefix(date, k) {
	// 		date = strings.Replace(date, k, v, 1)
	// 	}
	// }

	if date != "" {
		for _, v := range releaseDateFormats {
			t, err := time.Parse(v, date)
			if err == nil {
				return t.Unix()
			}
		}
	}

	return 0
}

func GetDaysToRelease(unix int64) string {

	release := time.Unix(unix, 0)

	days := math.Floor(release.Sub(now.BeginningOfDay()).Hours() / 24)

	if days == 0 {
		return "Today"
	}

	return "In " + GetTimeLong(int(days)*24*60, 2)
}
