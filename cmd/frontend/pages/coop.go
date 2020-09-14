package pages

import (
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gamedb/gamedb/cmd/frontend/helpers/datatable"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

const (
	maxPlayers = 10
)

func coopRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", coopHandler)
	r.Get("/search.json", coopSearchAjaxHandler)
	r.Get("/players.json", coopPlayersAjaxHandler)
	r.Get("/games.json", coopGames)
	r.Get("/{id}", coopHandler)
	return r
}

func coopHandler(w http.ResponseWriter, r *http.Request) {

	idStrings := helpers.StringToSlice(chi.URLParam(r, "id"), ",")
	idStrings = helpers.UniqueString(idStrings)

	if len(idStrings) > maxPlayers {
		returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "You can only compare games from up to " + strconv.Itoa(maxPlayers) + " people."})
		return
	}

	t := coopTemplate{}
	t.fill(w, r, "Co-op", "Find a game to play with friends.")
	t.IDs = idStrings

	returnTemplate(w, r, "coop", t)
}

type coopTemplate struct {
	globalTemplate
	IDs []string
}

func coopSearchAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := datatable.NewDataTableQuery(r, true)
	search := strings.TrimSpace(query.GetSearchString("search"))
	ids := helpers.StringToSlice(query.GetSearchString("ids"), ",")

	response := datatable.NewDataTablesResponse(r, query, 0, 0, nil)

	if search != "" {

		players, _, err := elasticsearch.SearchPlayers(5, 0, search, nil, nil)
		if err != nil {
			log.ErrS(err)
			return
		}

		for _, player := range players {

			var linkBool = helpers.SliceHasString(strconv.FormatInt(player.ID, 10), ids)
			var link = makeCoopActionLink(ids, strconv.FormatInt(player.ID, 10), linkBool)

			// Also update the response below
			response.AddRow([]interface{}{
				player.Games,              // 0
				player.ID,                 // 1
				player.GetName(),          // 2
				player.GetPath(),          // 3
				player.GetAvatar(),        // 4
				player.GetCommunityLink(), // 5
				player.Level,              // 6
				link,                      // 7
				linkBool,                  // 8
				player.Score,              // 9
				player.GetNameMarked(),    // 10
			})
		}
	}

	returnJSON(w, r, response)
}

func coopPlayersAjaxHandler(w http.ResponseWriter, r *http.Request) {

	query := datatable.NewDataTableQuery(r, true)
	ids := helpers.StringToSlice(query.GetSearchString("ids"), ",")
	ids2 := helpers.StringSliceToInt64Slice(ids)

	players, err := mongo.GetPlayersByID(ids2, bson.M{"_id": 1, "persona_name": 1, "avatar": 1, "level": 1, "games_count": 1})
	if err != nil {
		log.ErrS(err)
		return
	}

	var appMap = map[int64][]interface{}{}
	var response = datatable.NewDataTablesResponse(r, query, 0, 0, nil)
	for _, player := range players {

		var linkBool = helpers.SliceHasString(strconv.FormatInt(player.ID, 10), ids)
		var link = makeCoopActionLink(ids, strconv.FormatInt(player.ID, 10), linkBool)

		// Also update the response above
		appMap[player.ID] = []interface{}{
			player.GamesCount,      // 0
			player.ID,              // 1
			player.GetName(),       // 2
			player.GetPath(),       // 3
			player.GetAvatar(),     // 4
			player.CommunityLink(), // 5
			player.Level,           // 6
			link,                   // 7
			linkBool,               // 8
			0,                      // 9
			player.GetName(),       // 10
		}
	}

	// Looping here keeps the order the same as the URL IDs
	for _, v := range ids2 {
		if val, ok := appMap[v]; ok {
			response.AddRow(val)
		}
	}

	returnJSON(w, r, response)
}

func makeCoopActionLink(ids []string, id string, linkBool bool) string {

	var newIDs []string

	if linkBool {
		for _, v := range ids {
			if v != id {
				newIDs = append(newIDs, v)
			}
		}
	} else {
		newIDs = ids
		newIDs = append(newIDs, id)
	}

	return "/games/coop/" + strings.Join(newIDs, ",")
}

func coopGames(w http.ResponseWriter, r *http.Request) {

	query := datatable.NewDataTableQuery(r, true)
	ids := helpers.StringToSlice(query.GetSearchString("ids"), ",")

	var playerIDs []int64
	for _, v := range ids {
		i, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			i, err = helpers.IsValidPlayerID(i)
			if err == nil {
				playerIDs = append(playerIDs, i)
			}
		}
	}

	playerIDs = helpers.UniqueInt64(playerIDs)

	// Get players
	players, err := mongo.GetPlayersByID(playerIDs, bson.M{"_id": 1, "persona_name": 1, "avatar": 1})
	if err != nil {
		log.ErrS(err)
		returnErrorTemplate(w, r, errorTemplate{Code: 500})
		return
	}

	var foundPlayerIDs []int64
	for _, player := range players {
		foundPlayerIDs = append(foundPlayerIDs, player.ID)
	}

	// Queue players we dont already have
	for _, playerID := range playerIDs {
		if !helpers.SliceHasInt64(foundPlayerIDs, playerID) {

			ua := r.UserAgent()
			err = queue.ProducePlayer(queue.PlayerMessage{ID: playerID, UserAgent: &ua})
			if err == nil {
				log.Info("player queued", zap.String("ua", ua))
			}
			err = helpers.IgnoreErrors(err, queue.ErrIsBot, memcache.ErrInQueue)
			if err != nil {
				log.ErrS(err)
			}
		}
	}

	// Make a map of all games the players have
	var allApps = map[int]bool{}
	var allAppsByPlayer = map[int64][]int{}

	playerApps, err := mongo.GetPlayersApps(foundPlayerIDs, bson.M{"_id": 0, "player_id": 1, "app_id": 1})
	if err != nil {
		log.ErrS(err)
		return
	}

	for _, playerApp := range playerApps {

		allApps[playerApp.AppID] = true

		_, ok := allAppsByPlayer[playerApp.PlayerID]
		if ok {
			allAppsByPlayer[playerApp.PlayerID] = append(allAppsByPlayer[playerApp.PlayerID], playerApp.AppID)
		} else {
			allAppsByPlayer[playerApp.PlayerID] = []int{playerApp.AppID}
		}
	}

	// Remove apps that are not in a users apps
	for appID := range allApps {

		// Loop each user
		for _, gamesSlice := range allAppsByPlayer {

			if !helpers.SliceHasInt(gamesSlice, appID) {
				delete(allApps, appID)
				break
			}
		}
	}

	var count int64
	var games []coopGameTemplate

	if len(allApps) > 0 {

		// Convert to slice
		var appsSlice bson.A
		for appID := range allApps {
			appsSlice = append(appsSlice, appID)
		}

		filter := bson.D{
			// {"type", "game"},
			{"tags", bson.M{"$in": bson.A{128, 1685, 3843, 3841, 4508, 3859, 7368, 17770}}},
			{"_id", bson.M{"$in": appsSlice}},
		}

		projection := bson.M{"id": 1, "name": 1, "icon": 1, "platforms": 1, "achievements_count": 1, "tags": 1}

		var wg sync.WaitGroup

		wg.Add(1)
		go func() {

			defer wg.Done()

			apps, err := mongo.GetApps(query.GetOffset64(), 100, bson.D{{"reviews_score", -1}}, filter, projection)
			if err != nil {
				log.ErrS(err)
			}

			// Make visible tags
			for _, app := range apps {

				coopTags, err := app.GetCoopTags()
				if err != nil {
					log.ErrS(err)
					continue
				}

				games = append(games, coopGameTemplate{
					Game: app,
					Tags: coopTags,
				})
			}
		}()

		wg.Add(1)
		go func() {

			defer wg.Done()

			var err error
			count, err = mongo.CountDocuments(mongo.CollectionApps, filter, 60*60)
			if err != nil {
				log.ErrS(err)
			}
		}()

		wg.Wait()
	}

	response := datatable.NewDataTablesResponse(r, query, count, count, nil)

	for _, app := range games {

		response.AddRow([]interface{}{
			app.Game.ID,                  // 0
			app.Game.GetName(),           // 1
			app.Game.GetIcon(),           // 2
			app.Game.GetPlatformImages(), // 3
			app.Game.AchievementsCount,   // 4
			app.Tags,                     // 5
			app.Game.GetStoreLink(),      // 6
			app.Game.GetPath(),           // 7
		})
	}

	returnJSON(w, r, response)
}

type coopGameTemplate struct {
	Game mongo.App
	Tags string
}
