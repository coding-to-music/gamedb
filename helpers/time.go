package helpers

import (
	"math"
	"strconv"
	"strings"
	"time"
)

const (
	Day         = "02 Jan"
	DayTime     = "02 Jan 15:04"
	DayYearTime = "02 Jan 06 15:04"
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

type Periods struct {
	minutes string
	hours   string
	days    string
	months  string
	years   string
}

var (
	Short = Periods{"m", "h", "d", "m", "y"}
	Long  = Periods{" minutes", " hours", " days", " months", " years"}
)

func getHumanPlayTime(minutes int, periods Periods) (ret []string) {

	if minutes == 0 {
		return []string{"0" + periods.minutes}
	}

	then := time.Unix(int64(minutes)*60, 0)
	now := time.Unix(0, 0)

	_, years, months, days, hours, minutes, _, _ := Elapsed(then, now)

	var returns []string

	if years != 0 {
		returns = append(returns, strconv.Itoa(years)+periods.years)
	}
	if months != 0 {
		returns = append(returns, strconv.Itoa(months)+periods.months)
	}
	if days != 0 {
		returns = append(returns, strconv.Itoa(days)+periods.days)
	}
	if hours != 0 {
		returns = append(returns, strconv.Itoa(hours)+periods.hours)
	}
	if minutes != 0 {
		returns = append(returns, strconv.Itoa(minutes)+periods.minutes)
	}

	return returns
}

func GetTimeShort(minutes int, max int) (ret string) {
	returns := getHumanPlayTime(minutes, Short)

	x := int(math.Min(float64(max), float64(len(returns))))
	return strings.Join(returns[0:x], " ")
}

func GetTimeLong(minutes int, max int) (ret string) {
	returns := getHumanPlayTime(minutes, Long)

	x := int(math.Min(float64(max), float64(len(returns))))
	return strings.Join(returns[0:x], " ")
}
