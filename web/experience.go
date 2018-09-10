package web

import (
	"math"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/steam-authority/steam-authority/helpers"
)

const (
	totalRows = 5000
	chunkRows = 100
)

func ExperienceHandler(w http.ResponseWriter, r *http.Request) {

	var rows []level
	xp := 0

	for i := 0; i <= totalRows+1; i++ {

		rows = append(rows, level{
			Level: i,
			Start: xp,
		})

		xp = xp + int((math.Ceil((float64(i) + 1) / 10))*100)
	}

	rows[0] = level{
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

func chunk(rows []level, chunkSize int) (chunked [][]level) {

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
	Chunks      [][]level
	Level       int // Low value in form
	UserLevelTo int // High value in form
}

type level struct {
	Level int
	Start int
	End   int
	Diff  int
	Count int
}

func (l level) GetFriends() (ret int) {

	ret = 750

	if l.Level > 100 {
		ret = 1250
	}

	if l.Level > 200 {
		ret = 1750
	}

	if l.Level > 300 {
		ret = 2000
	}

	return ret
}

func (l level) GetAvatar2() string {
	return helpers.GetAvatar2(l.Level)
}
