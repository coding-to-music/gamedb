package pages

import (
	"net/http"
	"path"
	"sort"
	"strconv"
	"sync"

	"github.com/gamedb/gamedb/cmd/webserver/pages/helpers/datatable"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/i18n"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/gamedb/gamedb/pkg/tasks"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
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
	if len(country) > 4 {
		_, err := w.Write([]byte("invalid cc"))
		log.Err(err, r)
		return
	}

	var columns = map[string]string{
		"3": "level",
		"4": "badges_count",

		"5":  "games_count",
		"6":  "play_time",
		"13": "achievement_count",

		"7": "bans_game",
		"8": "bans_cav",
		"9": "bans_last",

		"10": "friends_count",
		"11": "comments_count",
	}

	var sortOrder = query.GetOrderMongo(columns)
	var filter = bson.D{}
	var isContinent bool

	// Continent
	for _, v := range i18n.Continents {
		if "c-"+v.Key == country {
			isContinent = true
			filter = append(filter, bson.E{Key: "continent_code", Value: v.Key})
			break
		}
	}

	// Country
	if !isContinent && country != "" {

		if country == "_" { // No country set
			country = ""
		}

		filter = append(filter, bson.E{Key: "country_code", Value: country})

		state := query.GetSearchString("state")
		if state != "" && len(state) <= 3 {
			filter = append(filter, bson.E{Key: "status_code", Value: state})
		}
	}

	search := query.GetSearchString("search")
	if len(search) >= 2 {

		search = path.Base(search) // Incase someone tries a profile URL

		filter = append(filter, bson.E{Key: "$text", Value: bson.M{"$search": search}})

		// quoted := regexp.QuoteMeta(search)
		// filter = append(filter, bson.E{Key: "$or", Value: bson.A{
		// 	bson.M{"persona_name": bson.M{"$regex": quoted, "$options": "i"}},
		// 	bson.M{"vanity_url": bson.M{"$regex": quoted, "$options": "i"}},
		// }})
	}

	//
	var wg sync.WaitGroup

	// Get players
	var players []mongo.Player
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		var projection = bson.M{
			"_id":          1,
			"persona_name": 1,
			"avatar":       1,
			"country_code": 1,
			//
			"level":        1,
			"badges_count": 1,
			//
			"games_count":       1,
			"play_time":         1,
			"achievement_count": 1,
			//
			"bans_game": 1,
			"bans_cav":  1,
			"bans_last": 1,
			//
			"friends_count":  1,
			"comments_count": 1,
		}

		players, err = mongo.GetPlayers(query.GetOffset64(), 100, sortOrder, filter, projection)
		if err != nil {
			log.Err(err, r)
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

	// Get filtered total
	var filtered int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		filtered, err = mongo.CountDocuments(mongo.CollectionPlayers, filter, 0)
		if err != nil {
			log.Err(err, r)
		}
	}()

	// Wait
	wg.Wait()

	var response = datatable.NewDataTablesResponse(r, query, total, filtered, nil)
	for k, v := range players {

		var index = query.GetOffset() + k + 1
		var id = strconv.FormatInt(v.ID, 10)
		var lastBan = v.LastBan.Format(helpers.DateYear)

		response.AddRow([]interface{}{
			index,                 // 0
			id,                    // 1
			v.GetName(),           // 2
			v.GetAvatar(),         // 3
			v.GetAvatar2(),        // 4
			v.Level,               // 5
			v.GamesCount,          // 6
			v.BadgesCount,         // 7
			v.GetPlaytimeShort(2), // 8
			v.GetPlaytimeShort(5), // 9
			v.FriendsCount,        // 10
			v.GetFlag(),           // 11
			v.GetCountry(),        // 12
			v.GetPath(),           // 13
			v.CommunityLink(),     // 14
			v.NumberOfGameBans,    // 15
			v.NumberOfVACBans,     // 16
			v.LastBan.Unix(),      // 17
			lastBan,               // 18
			v.CountryCode,         // 19
			v.CommentsCount,       // 20
			v.AchievementCount,    // 21
		})
	}

	returnJSON(w, r, response)
}
