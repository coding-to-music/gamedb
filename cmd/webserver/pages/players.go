package pages

import (
	"html/template"
	"net/http"
	"sort"
	"strconv"
	"sync"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/rounding"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/tasks"
	"github.com/go-chi/chi"
)

const CAF = "C-AF"
const CAN = "C-AN"
const CAS = "C-AS"
const CEU = "C-EU"
const CNA = "C-NA"
const CSA = "C-SA"
const COC = "C-OC"

type continent struct {
	Key   string
	Value string
}

// These strings must match the continents in the gountries library
var continents = []continent{
	{Key: CAF, Value: "Africa"},
	{Key: CAN, Value: "Antarctica"},
	{Key: CAS, Value: "Asia"},
	{Key: CEU, Value: "Australia"},
	{Key: CNA, Value: "Europe"},
	{Key: CSA, Value: "North America"},
	{Key: COC, Value: "South America"},
}

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
			log.Err(err, r)
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
		count, err = mongo.CountPlayers()
		log.Err(err, r)
	}()

	// Count players
	var countries []playersCountriesTemplate
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		codes, err := mongo.GetUniquePlayerCountries()
		if err != nil {
			log.Err(err)
			return
		}

		for _, v := range codes {
			countries = append(countries, playersCountriesTemplate{
				CC:   v,
				Name: helpers.CountryCodeToName(v),
			})
		}

		sort.Slice(countries, func(i, j int) bool {
			return countries[i].Name < countries[j].Name
		})
	}()

	// Wait
	wg.Wait()

	t := playersTemplate{}
	t.fill(w, r, "Players", "See where you come against the rest of the world ("+template.HTML(rounding.NearestThousandFormat(float64(count)))+" players).")
	t.Date = date
	t.Countries = countries
	t.Continents = continents

	returnTemplate(w, r, "players", t)
}

type playersTemplate struct {
	GlobalTemplate
	Date       string
	Countries  []playersCountriesTemplate
	Continents []continent
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

	var columns = map[string]string{
		"3": "level",
		"4": "games_count",
		"5": "badges_count",
		"6": "play_time",
		"7": "friends_count",

		"8":  "bans_game",
		"9":  "bans_cav",
		"10": "bans_last",
	}

	var sortOrder = query.getOrderMongo(columns, nil)
	var filter = mongo.M{}

	country := query.getSearchString("country")

	var isContinent bool
	for _, v := range continents {
		if v.Key == country {
			isContinent = true
			countriesIn := helpers.CountriesInContinent(v.Value)
			var countriesInA mongo.A
			for _, v := range countriesIn {
				countriesInA = append(countriesInA, v)
			}
			filter["country_code"] = mongo.M{"$in": countriesInA}
			break
		}
	}
	if !isContinent && country != "" {
		filter["country_code"] = country
	}

	state := query.getSearchString("state")
	if country == "US" && state != "" {
		filter["status_code"] = state
	}

	search := query.getSearchString("search")
	if len(search) >= 2 {
		sortOrder = nil
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

		var err error

		players, err := mongo.GetPlayers(query.getOffset64(), 100, sortOrder, filter, mongo.M{
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
			"bans_game":     1,
			"bans_cav":      1,
			"bans_last":     1,
			//
			"ranks": 1,
		})
		if err != nil {
			log.Err(err)
			return
		}

		for k, v := range players {

			playerRow := PlayerRow{}
			playerRow.Player = v

			switch query.getOrderString(columns) {
			case "badges_count":
				playerRow.setRank(v.GetRank(mongo.RankKeyBadges, mongo.RankCountryAll))
			case "friends_count":
				playerRow.setRank(v.GetRank(mongo.RankKeyFriends, mongo.RankCountryAll))
			case "games_count":
				playerRow.setRank(v.GetRank(mongo.RankKeyGames, mongo.RankCountryAll))
			case "level", "":
				playerRow.setRank(v.GetRank(mongo.RankKeyLevel, mongo.RankCountryAll))
			case "play_time":
				playerRow.setRank(v.GetRank(mongo.RankKeyPlaytime, mongo.RankCountryAll))
			case "comments":
				playerRow.setRank(v.GetRank(mongo.RankKeyComments, mongo.RankCountryAll))
			default:
				playerRow.Rank = strconv.Itoa(query.getOffset() + (k + 1))
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

	// Get filtered total
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

		response.AddRow([]interface{}{
			v.Rank,                                    // 0
			strconv.FormatInt(v.Player.ID, 10),        // 1
			v.Player.PersonaName,                      // 2
			v.Player.GetAvatar(),                      // 3
			v.Player.GetAvatar2(),                     // 4
			v.Player.Level,                            // 5
			v.Player.GamesCount,                       // 6
			v.Player.BadgesCount,                      // 7
			v.Player.GetTimeShort(),                   // 8
			v.Player.GetTimeLong(),                    // 9
			v.Player.FriendsCount,                     // 10
			v.Player.GetFlag(),                        // 11
			v.Player.GetCountry(),                     // 12
			v.Player.GetPath(),                        // 13
			v.Player.CommunityLink(),                  // 14
			v.Player.NumberOfGameBans,                 // 15
			v.Player.NumberOfVACBans,                  // 16
			v.Player.LastBan.Unix(),                   // 17
			v.Player.LastBan.Format(helpers.DateYear), // 18
			v.Player.CountryCode,                      // 19
		})
	}

	response.output(w, r)
}

// Rank struct
type PlayerRow struct {
	Player mongo.Player
	Rank   string
}

func (hp *PlayerRow) setRank(i int, found bool) {
	if found {
		hp.Rank = helpers.OrdinalComma(i)
	} else {
		hp.Rank = "-"
	}
}
