package helpers

import (
	"strconv"
	"time"
)

func DaysIn(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// https://stackoverflow.com/questions/36530251/golang-time-since-with-months-and-years
func Elapsed(from, to time.Time) (inverted bool, years, months, days, hours, minutes, seconds, nanoseconds int) {
	if from.Location() != to.Location() {
		to = to.In(to.Location())
	}

	inverted = false
	if from.After(to) {
		inverted = true
		from, to = to, from
	}

	y1, M1, d1 := from.Date()
	y2, M2, d2 := to.Date()

	h1, m1, s1 := from.Clock()
	h2, m2, s2 := to.Clock()

	ns1, ns2 := from.Nanosecond(), to.Nanosecond()

	years = y2 - y1
	months = int(M2 - M1)
	days = d2 - d1

	hours = h2 - h1
	minutes = m2 - m1
	seconds = s2 - s1
	nanoseconds = ns2 - ns1

	if nanoseconds < 0 {
		nanoseconds += 1e9
		seconds--
	}
	if seconds < 0 {
		seconds += 60
		minutes--
	}
	if minutes < 0 {
		minutes += 60
		hours--
	}
	if hours < 0 {
		hours += 24
		days--
	}
	if days < 0 {
		days += DaysIn(y2, M2-1)
		months--
	}
	if months < 0 {
		months += 12
		years--
	}

	return
}

func GetHumanPlayTime(minutes int) (ret string) {

	if minutes == 0 {
		return "0m"
	}

	// todo, only show the biggest 3 of the following

	then := time.Unix(int64(minutes)*60, 0)
	now := time.Unix(0, 0)

	_, years, months, days, hours, minutes, _, _ := Elapsed(then, now)

	if years != 0 {
		ret = ret + " " + strconv.Itoa(years) + "y"
	}
	if months != 0 {
		ret = ret + " " + strconv.Itoa(months) + "m"
	}
	if days != 0 {
		ret = ret + " " + strconv.Itoa(days) + "d"
	}
	if hours != 0 {
		ret = ret + " " + strconv.Itoa(hours) + "h"
	}
	if minutes != 0 {
		ret = ret + " " + strconv.Itoa(minutes) + "m"
	}

	return ret
}
