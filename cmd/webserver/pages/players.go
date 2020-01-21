package pages

import (
	"html/template"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"sync"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/tasks"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
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

	var wg sync.WaitGroup

	// Get config
	var date string
	wg.Add(1)
	go func() {

		defer wg.Done()

		config, err := tasks.GetTaskConfig(tasks.PlayerRanks{})
		if err != nil {
			err = helpers.IgnoreErrors(err, sql.ErrRecordNotFound)
			if err != nil {
				log.Err(err, r)
			}
		} else {
			date = config.Value
		}
	}()

	// Count players
	var count int64
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		count, err = mongo.CountDocuments(mongo.CollectionPlayers, nil, 0)
		if err != nil {
			log.Err(err, r)
		}
	}()

	// Get countries list
	var countries []playersCountriesTemplate
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		codes, err := mongo.GetUniquePlayerCountries()
		if err != nil {
			log.Err(err, r)
			return
		}

		for _, cc := range codes {

			// Add a star next to countries with states
			var star = ""
			if helpers.SliceHasString(mongo.CountriesWithStates, cc) {
				star = " *"
			}

			// Change value for empty country
			if cc == "" {
				cc = "_"
			}

			countries = append(countries, playersCountriesTemplate{
				CC:   cc,
				Name: helpers.CountryCodeToName(cc) + star,
			})
		}

		sort.Slice(countries, func(i, j int) bool {
			return countries[i].Name < countries[j].Name
		})
	}()

	// Get states
	states := map[string][]helpers.Tuple{}
	statesLock := sync.Mutex{}
	for _, cc := range mongo.CountriesWithStates {
		wg.Add(1)
		go func(cc string) {
			defer wg.Done()
			var err error
			statesLock.Lock()
			states[cc], err = mongo.GetUniquePlayerStates(cc)
			statesLock.Unlock()
			if err != nil {
				log.Err(err, r)
			}
		}(cc)
	}

	// Wait
	wg.Wait()

	t := playersTemplate{}
	t.fill(w, r, "Players", "See where you come against the rest of the world ("+template.HTML(helpers.ShortHandNumber(count))+" players).")
	t.Date = date
	t.Countries = countries
	t.Continents = helpers.Continents
	t.States = states

	returnTemplate(w, r, "players", t)
}

type playersTemplate struct {
	GlobalTemplate
	Date       string
	Countries  []playersCountriesTemplate
	Continents []helpers.Tuple
	States     map[string][]helpers.Tuple
}

type playersCountriesTemplate struct {
	CC   string
	Name string
}

func playersAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := DataTablesQuery{}
	err := query.fillFromURL(r.URL.Query())
	log.Err(err, r)

	query.limit(r)

	country := query.getSearchString("country")
	if len(country) > 4 {
		_, err = w.Write([]byte("invalid cc"))
		log.Err(err, r)
		return
	}

	var columns = map[string]string{
		"3": "level",
		"4": "badges_count",

		"5": "games_count",
		"6": "play_time",

		"7": "bans_game",
		"8": "bans_cav",
		"9": "bans_last",

		"10": "friends_count",
		"11": "comments_count",
	}

	var sortOrder = query.getOrderMongo(columns)
	var filter = bson.D{}
	var isContinent bool

	// Continent
	for _, v := range helpers.Continents {
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

		for _, cc := range mongo.CountriesWithStates {
			if cc == country {
				state := query.getSearchString(cc + "-state")
				if state != "" && len(state) <= 3 {
					filter = append(filter, bson.E{Key: "status_code", Value: state})
				}
			}
		}
	}

	search := query.getSearchString("search")
	if len(search) >= 2 {

		quoted := regexp.QuoteMeta(search)

		filter = append(filter, bson.E{Key: "$or", Value: bson.A{
			bson.M{"persona_name": bson.M{"$regex": quoted, "$options": "i"}},
			bson.M{"vanity_url": bson.M{"$regex": quoted, "$options": "i"}},
		}})
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
			"games_count": 1,
			"play_time":   1,
			//
			"bans_game": 1,
			"bans_cav":  1,
			"bans_last": 1,
			//
			"friends_count":  1,
			"comments_count": 1,
		}

		players, err = mongo.GetPlayers(query.getOffset64(), 100, sortOrder, filter, projection)
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

	response := DataTablesResponse{}
	response.RecordsTotal = total
	response.RecordsFiltered = filtered
	response.Draw = query.Draw
	response.limit(r)

	for k, v := range players {

		response.AddRow([]interface{}{
			helpers.OrdinalComma(query.getOffset() + k + 1), // 0
			strconv.FormatInt(v.ID, 10),                     // 1
			v.PersonaName,                                   // 2
			v.GetAvatar(),                                   // 3
			v.GetAvatar2(),                                  // 4
			v.Level,                                         // 5
			v.GamesCount,                                    // 6
			v.BadgesCount,                                   // 7
			v.GetTimeShort(),                                // 8
			v.GetTimeLong(),                                 // 9
			v.FriendsCount,                                  // 10
			v.GetFlag(),                                     // 11
			v.GetCountry(),                                  // 12
			v.GetPath(),                                     // 13
			v.CommunityLink(),                               // 14
			v.NumberOfGameBans,                              // 15
			v.NumberOfVACBans,                               // 16
			v.LastBan.Unix(),                                // 17
			v.LastBan.Format(helpers.DateYear),              // 18
			v.CountryCode,                                   // 19
			v.CommentsCount,                                 // 20
		})
	}

	response.output(w, r)
}
