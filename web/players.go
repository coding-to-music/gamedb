package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/Jleagle/go-helpers/logger"
	"github.com/go-chi/chi"
	slugify "github.com/gosimple/slug"
	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/mysql"
	"github.com/steam-authority/steam-authority/queue"
	"github.com/steam-authority/steam-authority/steam"
)

func RanksHandler(w http.ResponseWriter, r *http.Request) {

	// Normalise the order
	var ranks []datastore.Rank
	var err error

	switch chi.URLParam(r, "id") {
	case "badges":
		ranks, err = datastore.GetRanksBy("badges_rank")

		for k := range ranks {
			ranks[k].Rank = ranks[k].BadgesRank
		}
	case "friends":
		ranks, err = datastore.GetRanksBy("friends_rank")

		for k := range ranks {
			ranks[k].Rank = ranks[k].FriendsRank
		}
	case "games":
		ranks, err = datastore.GetRanksBy("games_rank")

		for k := range ranks {
			ranks[k].Rank = ranks[k].GamesRank
		}
	case "level", "":
		ranks, err = datastore.GetRanksBy("level_rank")

		for k := range ranks {
			ranks[k].Rank = ranks[k].LevelRank
		}
	case "playtime":
		ranks, err = datastore.GetRanksBy("play_time_rank")

		for k := range ranks {
			ranks[k].Rank = ranks[k].PlayTimeRank
		}
	default:
		err = errors.New("incorrect sort")
	}

	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 404, err.Error())
		return
	}

	// Count players
	playersCount, err := datastore.CountPlayers()
	if err != nil {
		logger.Error(err)
	}

	// Count ranks
	ranksCount, err := datastore.GetRanksCount()
	if err != nil {
		logger.Error(err)
	}

	template := playersTemplate{}
	template.Fill(r)
	template.Ranks = ranks
	template.PlayersCount = playersCount
	template.RanksCount = ranksCount

	returnTemplate(w, r, "ranks", template)
	return
}

type playersTemplate struct {
	GlobalTemplate
	Ranks        []datastore.Rank
	PlayersCount int
	RanksCount   int
}

func PlayerHandler(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	slug := chi.URLParam(r, "slug")

	idx, err := strconv.Atoi(id)
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 404, err.Error())
		return
	}

	player, err := datastore.GetPlayer(idx)
	if err != nil {
		logger.Error(err)
		returnErrorTemplate(w, r, 404, err.Error())
		return
	}

	err = player.UpdateIfNeeded()
	if err != nil {

		if err.Error() == steam.ErrorInvalidJson {
			returnErrorTemplate(w, r, 500, "Couldnt fetch player data, steam API may be down?")
			return
		}

		logger.Error(err)
		returnErrorTemplate(w, r, 500, err.Error())
		return
	}

	// Queue friends
	if player.ShouldUpdateFriends() {

		for _, v := range player.Friends {
			vv, _ := strconv.Atoi(v.SteamID)
			p, _ := json.Marshal(queue.PlayerMessage{
				PlayerID: vv,
			})
			queue.Produce(queue.PlayerQueue, p)
		}

		player.FriendsAddedAt = time.Now()
		player.Save()
	}

	// Redirect to correct slug
	correctSLug := slugify.Make(player.PersonaName)
	if slug != "" && slug != correctSLug {
		http.Redirect(w, r, "/players/"+id+"/"+correctSLug, 302)
		return
	}

	// Make friend ID slice
	var friendsSlice []int
	for _, v := range player.Friends {
		s, _ := strconv.Atoi(v.SteamID)
		friendsSlice = append(friendsSlice, s)
	}

	// Get friends for template
	friends, err := datastore.GetPlayersByIDs(friendsSlice)
	if err != nil {
		logger.Error(err)
	}

	// Get games
	var gamesSlice []int
	gamesMap := make(map[int]*playerAppTemplate)
	for _, v := range player.Games {
		gamesSlice = append(gamesSlice, v.AppID)
		gamesMap[v.AppID] = &playerAppTemplate{
			Time: v.PlaytimeForever,
		}
	}

	gamesSql, err := mysql.GetApps(gamesSlice, []string{"id", "name", "price_initial", "icon"})
	for _, v := range gamesSql {
		gamesMap[v.ID].ID = v.ID
		gamesMap[v.ID].Name = v.GetName()
		gamesMap[v.ID].Price = v.GetPriceInitial()
		gamesMap[v.ID].Icon = v.GetIcon()
	}

	// Sort games
	var sortedGamesSlice []*playerAppTemplate
	for _, v := range gamesMap {
		sortedGamesSlice = append(sortedGamesSlice, v)
	}

	sort.Slice(sortedGamesSlice, func(i, j int) bool {
		if sortedGamesSlice[i].Time == sortedGamesSlice[j].Time {
			return sortedGamesSlice[i].Name < sortedGamesSlice[j].Name
		}
		return sortedGamesSlice[i].Time > sortedGamesSlice[j].Time
	})

	// Template
	template := playerTemplate{}
	template.Fill(r)
	template.Player = player
	template.Friends = friends
	template.Games = sortedGamesSlice

	returnTemplate(w, r, "player", template)
}

type playerTemplate struct {
	GlobalTemplate
	Player  *datastore.Player
	Friends []datastore.Player
	Games   []*playerAppTemplate
}

type playerAppTemplate struct {
	ID    int
	Name  string
	Price string
	Icon  string
	Time  int
}

func (g playerAppTemplate) GetPriceHour() string {

	price, err := strconv.ParseFloat(g.Price, 64)
	if err != nil {
		price = 0
	}

	x := float64(price) / (float64(g.Time) / 60)
	if math.IsNaN(x) {
		x = 0
	}
	if math.IsInf(x, 0) {
		return "∞"
	}
	return fmt.Sprintf("%0.2f", x)
}

func PlayerIDHandler(w http.ResponseWriter, r *http.Request) {

	post := r.PostFormValue("id")

	// todo, check DB before doing api call

	id, err := steam.GetID(post)
	if err != nil {
		logger.Info(err.Error() + ": " + post)
		returnErrorTemplate(w, r, 404, "Can't find user: "+post)
		return
	}

	http.Redirect(w, r, "/players/"+id, 302)
}
