package pages

import (
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/datatable"
	elastic_search "github.com/gamedb/gamedb/pkg/elastic-search"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/i18n"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/tasks"
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
	var date string
	wg.Add(1)
	go func() {

		defer wg.Done()

		config, err := tasks.GetTaskConfig(tasks.PlayersUpdateRanks{})
		if err != nil {
			err = helpers.IgnoreErrors(err, mysql.ErrRecordNotFound)
			if err != nil {
				log.Err(err, r)
			}
		} else {
			date = config.Value
		}
	}()

	// Get countries list
	var countries []playersCountriesTemplate
	wg.Add(1)
	go func() {

		defer wg.Done()

		for cc := range i18n.States {

			if cc == "" {
				cc = "_"
			}

			countries = append(countries, playersCountriesTemplate{
				CC:   cc,
				Name: i18n.CountryCodeToName(cc),
			})
		}

		sort.Slice(countries, func(i, j int) bool {
			return countries[i].Name < countries[j].Name
		})
	}()

	// Wait
	wg.Wait()

	t := playersTemplate{}
	t.fill(w, r, "Players", "See where you come against the rest of the world")
	t.Date = date
	t.Countries = countries

	returnTemplate(w, r, "players", t)
}

type playersTemplate struct {
	GlobalTemplate
	Date      string
	Countries []playersCountriesTemplate
}

type playersCountriesTemplate struct {
	CC   string
	Name string
}

func statesAjaxHandler(w http.ResponseWriter, r *http.Request) {

	cc := r.URL.Query().Get("cc")

	var states []helpers.Tuple

	if val, ok := i18n.States[cc]; ok {

		for k, v := range val {
			states = append(states, helpers.Tuple{Key: k, Value: v})
		}

		sort.Slice(states, func(i, j int) bool {
			return states[i].Value < states[j].Value
		})
	}

	returnJSON(w, r, states)
}

func playersAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := datatable.NewDataTableQuery(r, true)

	country := query.GetSearchString("country")
	state := query.GetSearchString("state")
	search := query.GetSearchString("search")

	var sorters = query.GetOrderElastic(map[string]string{
		"3": "level",
		"4": "badges",

		"5":  "games",
		"6":  "play_time",
		"13": "achievements",

		"7": "game_bans",
		"8": "vac_bans",
		"9": "last_ban",

		"10": "friends",
		"11": "comments",
	})

	var isContinent bool

	var filters []elastic.Query

	if country != "" {

		for _, v := range i18n.Continents {
			if "c-"+v.Key == country {

				isContinent = true
				filters = append(filters, elastic.NewTermQuery("continent", country))
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
	var players []elastic_search.Player
	var filtered int64
	var aggregations = map[string]map[string]int64{}
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		var aggs elastic.Aggregations

		players, aggs, filtered, err = elastic_search.SearchPlayers(100, query.GetOffset(), search, sorters, filters)
		if err != nil {
			log.Err(err, r)
			return
		}

		if a, ok := aggs.Terms("country"); ok {
			aggregations["country"] = map[string]int64{}
			for _, v := range a.Buckets {
				aggregations["country"][v.Key.(string)] = v.DocCount
				if !isContinent && country != "" && country == v.Key.(string) {
					if a, ok := v.Terms("state"); ok {
						aggregations["state"] = map[string]int64{}
						for _, v := range a.Buckets {
							aggregations["state"][v.Key.(string)] = v.DocCount
						}
					}
				}
			}
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
			log.Err(err, r)
		}
	}()

	// Wait
	wg.Wait()

	var response = datatable.NewDataTablesResponse(r, query, total, filtered, aggregations)
	for k, v := range players {

		var index = query.GetOffset() + k + 1
		var id = strconv.FormatInt(v.ID, 10)
		var lastBan = time.Unix(v.LastBan, 0).Format(helpers.DateYear)

		var playtimeShort = helpers.GetTimeShort(v.PlayTime, 2)
		var playtimeLong = helpers.GetTimeShort(v.PlayTime, 5)

		response.AddRow([]interface{}{
			index,             // 0
			id,                // 1
			v.GetName(),       // 2
			v.GetAvatar(),     // 3
			v.GetAvatar2(),    // 4
			v.Level,           // 5
			v.Games,           // 6
			v.Badges,          // 7
			playtimeShort,     // 8
			playtimeLong,      // 9
			v.Friends,         // 10
			v.GetFlag(),       // 11
			v.GetCountry(),    // 12
			v.GetPath(),       // 13
			v.CommunityLink(), // 14
			v.GameBans,        // 15
			v.VACBans,         // 16
			v.LastBan,         // 17
			lastBan,           // 18
			v.CountryCode,     // 19
			v.Comments,        // 20
			v.Achievements,    // 21
		})
	}

	returnJSON(w, r, response)
}
