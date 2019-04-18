package pages

import (
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gamedb/website/pkg/helpers"
	"github.com/gamedb/website/pkg/log"
	"github.com/go-chi/chi"
)

const (
	totalRows = 5000
	chunkRows = 100
)

func experienceRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", experienceHandler)
	r.Get("/{id}", experienceHandler)
	return r
}

func experienceHandler(w http.ResponseWriter, r *http.Request) {

	ret := setAllowedQueries(w, r, []string{})
	if ret {
		return
	}

	setCacheHeaders(w, time.Hour*24)

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
	t.fill(w, r, "Experience", "Check how much XP you need to go up a level")
	t.Chunks = chunkExperienceRow(rows, chunkRows)

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

	err := returnTemplate(w, r, "experience", t)
	log.Err(err, r)
}

func chunkExperienceRow(rows []level, chunkSize int) (chunked [][]level) {

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
	Chunks [][]level
	Level  int // Low value in form
}

type level struct {
	Level int
	Start int
	End   int
	Diff  int
	Count int
}

func (l level) GetFriends() int {
	return helpers.GetPlayerMaxFriends(l.Level)
}

func (l level) GetAvatar2() string {
	return helpers.GetAvatar2(l.Level)
}
