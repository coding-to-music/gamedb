package pages

import (
	"html/template"
	"net/http"
	"sync"

	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/go-chi/chi"
)

func PlayersRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", playersHandler)
	r.Get("/add", playerAddHandler)
	r.Post("/add", playerAddHandler)
	r.Get("/players.json", playersAjaxHandler)
	r.Mount("/{id:[0-9]+}", PlayerRouter())
	return r
}

func playersHandler(w http.ResponseWriter, r *http.Request) {

	// Template
	t := playersTemplate{}
	t.setRandomBackground()

	//
	var wg sync.WaitGroup

	// Get config
	wg.Add(1)
	go func() {

		defer wg.Done()

		config, err := sql.GetConfig(sql.ConfRanksUpdated)
		log.Err(err, r)

		if err == nil {
			t.Date = config.Value
		}

	}()

	// Count players
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.PlayersCount, err = mongo.CountPlayers()
		log.Err(err, r)

	}()

	// Wait
	wg.Wait()

	t.fill(w, r, "Players", "See where you come against the rest of the world ("+template.HTML(humanize.Comma(t.PlayersCount))+" players).")

	err := returnTemplate(w, r, "players", t)
	log.Err(err, r)
}

type playersTemplate struct {
	GlobalTemplate
	PlayersCount int64
	Date         string
}

func playersAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	query.limit(r)

	var columns = map[string]string{
		"3": "level",
		"4": "games_count",
		"5": "badges_count",
		"6": "play_time",
		"7": "friends_count",
	}

	var sort = query.getOrderMongo(columns, nil)
	var filter = mongo.M{}

	search := query.getSearchString("search")
	if len(search) >= 2 {
		sort = nil
		filter["$or"] = mongo.A{
			mongo.M{"$text": mongo.M{"$search": search}},
			mongo.M{"_id": search},
		}
	}

	//
	var wg sync.WaitGroup

	// Get players
	var playerRows []PlayerRow
	wg.Add(1)
	go func() {

		defer wg.Done()

		players, err := mongo.GetPlayers(query.getOffset64(), 100, sort, filter, mongo.M{
			"_id":          1,
			"persona_name": 1,
			"avatar":       1,
			"country_code": 1,
			//
			"badges_count":  1,
			"friends_count": 1,
			"games_count":   1,
			"level":         1,
			"play_time":     1,
			//
			"badges_rank":    1,
			"friends_rank":   1,
			"games_rank":     1,
			"level_rank":     1,
			"play_time_rank": 1,
		}, nil)
		if err != nil {
			log.Err(err)
			return
		}

		for _, v := range players {

			playerRow := PlayerRow{}
			playerRow.Player = v

			switch query.getOrderString(columns) {
			case "badges_count":
				playerRow.Rank = v.BadgesRank
			case "friends_count":
				playerRow.Rank = v.FriendsRank
			case "games_count":
				playerRow.Rank = v.GamesRank
			case "level", "":
				playerRow.Rank = v.LevelRank
			case "play_time":
				playerRow.Rank = v.PlayTimeRank
			}

			playerRows = append(playerRows, playerRow)
		}
	}()

	// Get total
	var total int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		total, err = mongo.CountPlayers()
		log.Err(err, r)
	}()

	// Get total
	var filtered int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		filtered, err = mongo.CountDocuments(mongo.CollectionPlayers, filter, 0)
		log.Err(err, r)
	}()

	// Wait
	wg.Wait()

	response := DataTablesAjaxResponse{}
	response.RecordsTotal = total
	response.RecordsFiltered = filtered
	response.Draw = query.Draw
	response.limit(r)

	for _, v := range playerRows {

		response.AddRow(v.Player.OutputForJSON(v.GetRank()))
	}

	response.output(w, r)
}

type PlayerRow struct {
	Player mongo.Player
	Rank   int
}

func (pr PlayerRow) GetRank() string {

	if pr.Rank == 0 {
		return "-"
	}

	return helpers.OrdinalComma(pr.Rank)
}
