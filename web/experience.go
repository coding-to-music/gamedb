package web

import (
	"math"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

const (
	totalRows = 3000
	chunkRows = 100
)

func ExperienceHandler(w http.ResponseWriter, r *http.Request) {

	var rows []experienceRow
	xp := 0

	for i := 0; i <= totalRows+1; i++ {

		rows = append(rows, experienceRow{
			Level: i,
			Start: int(xp),
		})

		xp = xp + int((math.Ceil((float64(i) + 1) / 10))*100)
	}

	rows[0] = experienceRow{
		Level: 0,
		End:   99,
		Diff:  100,
	}

	for i := 1; i <= totalRows; i++ {

		thisRow := rows[i]
		nextRow := rows[i+1]

		rows[i].Diff = nextRow.Start - thisRow.Start
		rows[i].End = nextRow.Start - 1
	}

	rows = rows[0 : totalRows+1]

	t := experienceTemplate{}
	t.Fill(w, r, "Experience")
	t.Chunks = chunk(rows, chunkRows)

	// Default calculator levels if logged out
	t.UserLevelTo = t.UserLevel + 10
	if t.UserLevel == 0 {
		t.UserLevel = 10
		t.UserLevelTo = 20
	}

	// Highlight level from URL
	t.Level = -1
	id := chi.URLParam(r, "id")
	if id != "" {
		i, err := strconv.Atoi(id)
		if err != nil {
			t.Level = -1
		} else {
			t.Level = i
		}

	}

	returnTemplate(w, r, "experience", t)
}

func chunk(rows []experienceRow, chunkSize int) (chunked [][]experienceRow) {

	for i := 0; i < len(rows); i += chunkSize {
		end := i + chunkSize

		if end > len(rows) {
			end = len(rows)
		}

		chunked = append(chunked, rows[i:end])
	}

	return chunked
}

type experienceTemplate struct {
	GlobalTemplate
	Chunks      [][]experienceRow
	Level       int
	UserLevelTo int
}

type experienceRow struct {
	Level int
	Start int
	End   int
	Diff  int
	Count int
}

func (e experienceRow) GetFriends() (ret int) {

	ret = 750

	if e.Level > 100 {
		ret = 1250
	}

	if e.Level > 200 {
		ret = 1750
	}

	if e.Level > 300 {
		ret = 2000
	}

	return ret
}
