package pages

import (
	"net/http"
	"strconv"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/go-chi/chi"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	maxPlayers = 10
)

func CoopRouter() http.Handler {

	r := chi.NewRouter()
	r.Get("/", coopHandler)
	return r
}

func coopHandler(w http.ResponseWriter, r *http.Request) {

	t := coopTemplate{}
	t.fill(w, r, "Co-op", "Find a game to play with friends.")
	t.DefaultAvatar = helpers.DefaultAppIcon

	// Get player ints
	var playerIDs []int64
	for _, v := range r.URL.Query()["p"] {
		i, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			i, err = helpers.IsValidPlayerID(i)
			if err == nil {
				playerIDs = append(playerIDs, i)
			}
		}
	}

	playerIDs = helpers.UniqueInt64(playerIDs)

	// Check for max number of players
	if len(playerIDs) > maxPlayers {
		returnErrorTemplate(w, r, errorTemplate{Code: 404, Message: "You can only compare games from up to " + strconv.Itoa(maxPlayers) + " people."})
		return
	}

	// Get players
	var err error
	t.Players, err = mongo.GetPlayersByID(playerIDs, bson.M{"_id": 1, "persona_name": 1, "avatar": 1})
	if err != nil {
		log.Err(r, err)
		returnErrorTemplate(w, r, errorTemplate{Code: 500})
		return
	}

	var foundPlayerIDs []int64
	for _, player := range t.Players {
		foundPlayerIDs = append(foundPlayerIDs, player.ID)
	}

	// Queue players we dont already have
	for _, playerID := range playerIDs {
		if !helpers.SliceHasInt64(foundPlayerIDs, playerID) {

			ua := r.UserAgent()
			err = queue.ProducePlayer(queue.PlayerMessage{ID: playerID, UserAgent: &ua})
			if err == nil {
				log.Info(log.LogNameTriggerUpdate, r, ua)
			}
			err = helpers.IgnoreErrors(err, queue.ErrIsBot, memcache.ErrInQueue)
			if err != nil {
				log.Err(err)
			} else {
				t.addToast(Toast{Title: "Update", Message: "Player has been queued for an update"})
			}
		}
	}

	// Make a map of all games the players have
	var allApps = map[int]bool{}
	var allAppsByPlayer = map[int64][]int{}

	playerApps, err := mongo.GetPlayersApps(foundPlayerIDs, bson.M{"_id": -1, "player_id": 1, "app_id": 1})
	log.Err(err)
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

		var remove = false

		// Loop each user
		for _, gamesSlice := range allAppsByPlayer {

			if !helpers.SliceHasInt(gamesSlice, appID) {
				remove = true
				break
			}
		}

		if remove {
			delete(allApps, appID)
		}
	}

	if len(allApps) > 0 {

		// Convert to slice
		var appsSlice bson.A
		for appID := range allApps {
			appsSlice = append(appsSlice, appID)
		}

		filter := bson.D{
			{"type", "game"},
			{"tags", bson.M{"$in": bson.A{128, 1685, 3843, 3841, 4508, 3859, 7368, 17770}}},
			{"_id", bson.M{"$in": appsSlice}},
		}

		projection := bson.M{"id": 1, "name": 1, "icon": 1, "platforms": 1, "achievements": 1, "tags": 1}

		apps, err := mongo.GetApps(0, 500, bson.D{{"reviews_score", 1}}, filter, projection, nil)
		if err != nil {
			log.Err(err, r)
		}

		// Make visible tags
		for _, app := range apps {

			coopTags, err := app.GetCoopTags()
			log.Err(err, r)

			t.Games = append(t.Games, coopGameTemplate{
				Game: app,
				Tags: coopTags,
			})
		}
	}

	returnTemplate(w, r, "coop", t)
}

type coopTemplate struct {
	GlobalTemplate
	Players       []mongo.Player
	Games         []coopGameTemplate
	DefaultAvatar string
}

type coopGameTemplate struct {
	Game mongo.App
	Tags string
}
