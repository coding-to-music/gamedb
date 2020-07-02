package pages

import (
	"encoding/gob"
	"math"
	"net/http"
	"strconv"

	sessionHelpers "github.com/gamedb/gamedb/cmd/webserver/pages/helpers/session"
	"github.com/gamedb/gamedb/pkg/cache"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
)

const (
	totalRows = 5000
	chunkRows = 100
)

func init() {
	gob.Register(&[][]level{})
}

func ExperienceRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", experienceHandler)
	r.Get("/{id}", experienceHandler)
	return r
}

func experienceHandler(w http.ResponseWriter, r *http.Request) {

	t := experienceTemplate{}
	t.fill(w, r, "Experience", "Check how much XP you need to go up a level")

	//
	var chunks [][]level

	retrieve := func() (interface{}, error) {
		chunks = getExperienceRows()
		return chunks, nil
	}

	err := cache.GetSetCache("experience", 0, retrieve, &chunks)
	if err != nil {
		log.Err(err, r)
		returnErrorTemplate(w, r, errorTemplate{Code: 500})
	}

	t.Chunks = chunks

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

	level := sessionHelpers.Get(r, sessionHelpers.SessionPlayerLevel)
	if level == "" {
		t.PlayerLevel = 10
		t.PlayerLevelTo = 20
	} else {
		t.PlayerLevel, err = strconv.Atoi(level)
		if err != nil {
			log.Err(err, r)
		}
		t.PlayerLevelTo = t.PlayerLevel + 10
	}

	returnTemplate(w, r, "experience", t)
}

func getExperienceRows() (chunked [][]level) {

	levels, err := mongo.GetPlayerLevels()
	if err != nil {
		log.Err(err)
		return
	}

	var levelsMap = map[int]int{}
	for _, v := range levels {
		levelsMap[v.ID] = v.Count
	}

	//noinspection GoPreferNilSlice
	var rows = []level{}
	xp := 0

	for i := 0; i <= totalRows+1; i++ {

		var row = level{
			Level: i,
			Start: xp,
		}

		if val, ok := levelsMap[i]; ok {
			row.Players = val
		}

		rows = append(rows, row)

		xp = xp + int((math.Ceil((float64(i)+1)/10))*100)
	}

	for i := 1; i <= totalRows; i++ {

		thisRow := rows[i]
		nextRow := rows[i+1]

		rows[i].Diff = nextRow.Start - thisRow.Start
		rows[i].End = nextRow.Start - 1
	}

	rows = rows[0 : totalRows+1]

	// Chunk
	for i := 0; i < len(rows); i += chunkRows {
		end := i + chunkRows

		if end > len(rows) {
			end = len(rows)
		}

		chunked = append(chunked, rows[i:end])
	}

	return chunked
}

type experienceTemplate struct {
	GlobalTemplate
	Chunks        [][]level
	Level         int // Low value in form
	PlayerLevel   int
	PlayerLevelTo int
}

type level struct {
	Level   int
	Start   int
	End     int
	Diff    int
	Count   int
	Players int
}

func (l level) GetFriends() int {
	return helpers.GetPlayerMaxFriends(l.Level)
}

func (l level) GetAvatar2() string {
	return helpers.GetPlayerAvatar2(l.Level)
}
