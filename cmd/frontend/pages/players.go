package pages

import (
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/cmd/frontend/helpers/datatable"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/i18n"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/go-chi/chi"
	"github.com/olivere/elastic/v7"
)

func PlayersRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", playersHandler)
	r.Get("/add", playerAddHandler)
	r.Post("/add", playerAddHandler)
	r.Get("/states.json", statesAjaxHandler)
	r.Get("/players.json", playersAjaxHandler)
	r.Mount("/{id:[0-9]+}", PlayerRouter())
	return r
}

func playersHandler(w http.ResponseWriter, r *http.Request) {

	var wg sync.WaitGroup

	// Get config
	// var date string
	// wg.Add(1)
	// go func() {
	//
	// 	defer wg.Done()
	//
	// 	config, err := tasks.GetTaskConfig(tasks.PlayersUpdateRanks{})
	// 	if err != nil {
	// 		err = helpers.IgnoreErrors(err, mysql.ErrRecordNotFound)
	// 		if err != nil {
	// 			log.ErrS(err)
	// 		}
	// 	} else {
	// 		date = config.Value
	// 	}
	// }()

	var total int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		total, err = mongo.CountDocuments(mongo.CollectionPlayers, nil, 0)
		if err != nil {
			log.ErrS(err)
		}
	}()

	// Get countries list
	var countries []helpers.Tuple
	var continents []i18n.Continent
	wg.Add(1)
	go func() {

		defer wg.Done()

		aggs, err := elasticsearch.AggregatePlayerCountries()
		if err != nil {
			log.ErrS(err)
		}

		for cc := range i18n.States {

			var countryAgg string
			if val, ok := aggs[cc]; ok {
				countryAgg = " (" + humanize.Comma(val) + ")"
			}

			countries = append(countries, helpers.Tuple{
				Key:   cc,
				Value: i18n.CountryCodeToName(cc) + countryAgg,
			})
		}

		sort.Slice(countries, func(i, j int) bool {
			return countries[i].Value < countries[j].Value
		})

		var noCountryAgg string
		if val, ok := aggs[""]; ok {
			noCountryAgg = " (" + humanize.Comma(val) + ")"
		}

		// Prepend
		countries = append(
			[]helpers.Tuple{
				// {Key: "", Value: "All Countries (" + humanize.Comma(total) + ")"},
				{Key: "_", Value: "No Country" + noCountryAgg},
				// {Key: "", Value: "---"},
			},
			countries...
		)

		// Continents
		continents = append(continents, i18n.Continents...) // Copy without reference

		for k, v := range continents {
			if val, ok := aggs["c-"+v.Key]; ok {
				continents[k].Value += " (" + humanize.Comma(val) + ")"
			}
		}
	}()

	// Wait
	wg.Wait()

	t := playersTemplate{}
	t.fill(w, r, "Players", "See where you come against the rest of the world")
	t.Countries = countries
	t.Continents2 = continents
	t.Total = total

	returnTemplate(w, r, "players", t)
}

type playersTemplate struct {
	globalTemplate
	Countries   []helpers.Tuple
	Continents2 []i18n.Continent
	Total       int64
}

func (t playersTemplate) includes() []string {
	return []string{"includes/players_header.gohtml"}
}

func statesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	//noinspection GoPreferNilSlice
	var states = []helpers.Tuple{}

	cc := r.URL.Query().Get("cc")
	if cc != "" {

		aggs, err := elasticsearch.AggregatePlayerCountries()
		if err != nil {
			log.ErrS(err)
		}

		if val, ok := i18n.States[cc]; ok {

			for k, v := range val {

				var stateAgg string
				if val, ok := aggs[cc+"-"+k]; ok {
					stateAgg = " (" + humanize.Comma(val) + ")"
				}

				states = append(states, helpers.Tuple{Key: k, Value: v + stateAgg})
			}

			sort.Slice(states, func(i, j int) bool {
				return states[i].Value < states[j].Value
			})

			var countryAgg string
			if val, ok := aggs[cc]; ok {
				countryAgg = " (" + humanize.Comma(val) + ")"
			}

			var stateAgg2 string
			if val, ok := aggs[cc+"-"]; ok {
				stateAgg2 = " (" + humanize.Comma(val) + ")"
			}

			// Prepend
			states = append(
				[]helpers.Tuple{
					{Key: "", Value: "All States" + countryAgg},
					{Key: "_", Value: "No State" + stateAgg2},
					{Key: "", Value: "---"},
				},
				states...
			)
		}
	}

	returnJSON(w, r, states)
}

func playersAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := datatable.NewDataTableQuery(r, true)

	country := query.GetSearchString("country")
	state := query.GetSearchString("state")
	search := query.GetSearchString("search")

	var sorters = query.GetOrderElastic(map[string]string{
		"3":  "level",
		"4":  "badges",
		"14": "badges_foil",

		"5": "games",
		"6": "play_time",

		"7": "game_bans",
		"8": "vac_bans",
		"9": "last_ban",

		"10": "friends",
		"11": "comments",

		"12": "achievements",
		"13": "achievements_100",
	})

	var isContinent bool

	var filters []elastic.Query

	if country != "" {

		for _, v := range i18n.Continents {
			if "c-"+v.Key == country {

				isContinent = true
				filters = append(filters, elastic.NewTermQuery("continent", v.Key))
				break
			}
		}

		if !isContinent {

			if _, ok := i18n.States[country]; ok || country == "_" {

				if country == "_" {
					country = ""
				}

				filters = append(filters, elastic.NewTermQuery("country_code", country))

				if _, ok := i18n.States[country][state]; ok || state == "_" {

					if state == "_" {
						state = ""
					}

					filters = append(filters, elastic.NewTermQuery("state_code", state))
				}
			}
		}
	}

	//
	var wg sync.WaitGroup

	// Get players
	var players []elasticsearch.Player
	var filtered int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		players, filtered, err = elasticsearch.SearchPlayers(100, query.GetOffset(), search, sorters, filters)
		if err != nil {
			log.ErrS(err)
		}
	}()

	// Get total
	var total int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		total, err = mongo.CountDocuments(mongo.CollectionPlayers, nil, 0)
		if err != nil {
			log.ErrS(err)
		}
	}()

	// Wait
	wg.Wait()

	var response = datatable.NewDataTablesResponse(r, query, total, filtered, nil)
	for k, v := range players {

		var index = query.GetOffset() + k + 1
		var id = strconv.FormatInt(v.ID, 10)
		var lastBan = time.Unix(v.LastBan, 0).Format(helpers.DateYear)

		var playtimeShort = helpers.GetTimeShort(v.PlayTime, 2)
		var playtimeLong = helpers.GetTimeShort(v.PlayTime, 5)

		response.AddRow([]interface{}{
			index,                // 0
			id,                   // 1
			v.GetName(),          // 2
			v.GetAvatar(),        // 3
			v.GetAvatar2(),       // 4
			v.Level,              // 5
			v.Games,              // 6
			v.Badges,             // 7
			playtimeShort,        // 8
			playtimeLong,         // 9
			v.Friends,            // 10
			v.GetFlag(),          // 11
			v.GetCountry(),       // 12
			v.GetPath(),          // 13
			v.GetCommunityLink(), // 14
			v.GameBans,           // 15
			v.VACBans,            // 16
			v.LastBan,            // 17
			lastBan,              // 18
			v.CountryCode,        // 19
			v.Comments,           // 20
			v.Achievements,       // 21
			v.Achievements100,    // 22
			v.GetNameMarked(),    // 23
			v.Score,              // 24
			v.BadgesFoil,         // 25
		})
	}

	returnJSON(w, r, response)
}
