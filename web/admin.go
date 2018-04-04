package web

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi"
	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
	"github.com/steam-authority/steam-authority/queue"
	"github.com/steam-authority/steam-authority/steam"
)

func AdminHandler(w http.ResponseWriter, r *http.Request) {

	option := chi.URLParam(r, "option")

	switch option {
	case "re-add-all-apps":
		go adminApps()
	case "deploy":
		go adminDeploy()
	case "count-donations":
		go adminDonations()
	case "count-genres":
		go adminGenres()
	case "queues":
		r.ParseForm()
		go adminQueues(r)
	case "calculate-ranks":
		go adminRanks()
	case "count-tags":
		go adminTags()
	case "disable-consumers":
		go adminDisableConsumers()
	}

	// Redirect away after action
	if option != "" {
		http.Redirect(w, r, "/admin?"+option, 302)
		return
	}

	// Get configs for times
	configs, err := datastore.GetMultiConfigs([]string{
		datastore.ConfTagsUpdated,
		datastore.ConfGenresUpdated,
		datastore.ConfGenresUpdated,
		datastore.ConfDonationsUpdated,
		datastore.ConfRanksUpdated,
		datastore.ConfAddedAllApps,
	})
	if err != nil {
		logger.Error(err)
	}

	// Template
	template := adminTemplate{}
	template.Fill(r, "Admin")
	template.Configs = configs

	returnTemplate(w, r, "admin", template)
	return
}

type adminTemplate struct {
	GlobalTemplate
	Errors  []string
	Configs map[string]datastore.Config
}

func adminDisableConsumers() {

}

func adminApps() {

	// Get apps
	apps, err := steam.GetAppList()
	if err != nil {
		logger.Error(err)
		return
	}

	for _, v := range apps {
		bytes, _ := json.Marshal(queue.AppMessage{
			AppID:    v.AppID,
			ChangeID: 0,
		})

		queue.Produce(queue.AppQueue, bytes)
	}

	//
	err = datastore.SetConfig(datastore.ConfAddedAllApps, strconv.Itoa(int(time.Now().Unix())))
	if err != nil {
		logger.Error(err)
	}

	logger.Info(strconv.Itoa(len(apps)) + " apps added to rabbit")
}

func adminDeploy() {

	//
	err := datastore.SetConfig(datastore.ConfDeployed, strconv.Itoa(int(time.Now().Unix())))
	if err != nil {
		logger.Error(err)
	}
}

func adminDonations() {

	donations, err := datastore.GetDonations(0, 0)
	if err != nil {
		logger.Error(err)
		return
	}

	// map[player]total
	counts := make(map[int]int)

	for _, v := range donations {

		if _, ok := counts[v.PlayerID]; ok {
			counts[v.PlayerID] = counts[v.PlayerID] + v.AmountUSD
		} else {
			counts[v.PlayerID] = v.AmountUSD
		}
	}

	for k, v := range counts {
		player, err := datastore.GetPlayer(k)
		if err != nil {
			logger.Error(err)
			continue
		}

		player.Donated = v
		_, err = datastore.SaveKind(player.GetKey(), player)
	}

	//
	err = datastore.SetConfig(datastore.ConfDonationsUpdated, strconv.Itoa(int(time.Now().Unix())))
	if err != nil {
		logger.Error(err)
	}

	logger.Info("Updated " + strconv.Itoa(len(counts)) + " player donation counts")
}

// todo, handle genres that no longer have any games.
func adminGenres() {

	filter := url.Values{}
	filter.Set("genres_depth", "3")

	apps, err := mysql.SearchApps(filter, 0, "", []string{})
	if err != nil {
		logger.Error(err)
	}

	counts := make(map[int]*adminGenreCount)

	for _, app := range apps {
		genres, err := app.GetGenres()
		if err != nil {
			logger.Error(err)
			continue
		}

		for _, genre := range genres {
			if _, ok := counts[genre.ID]; ok {
				counts[genre.ID].Count++
			} else {
				counts[genre.ID] = &adminGenreCount{
					Count: 1,
					Genre: genre,
				}
			}
		}
	}

	var wg sync.WaitGroup

	for _, v := range counts {

		wg.Add(1)
		go func(v *adminGenreCount) {

			err := mysql.SaveOrUpdateGenre(v.Genre.ID, v.Genre.Description, v.Count)
			if err != nil {
				logger.Error(err)
			}

			wg.Done()

		}(v)
	}
	wg.Wait()

	//
	err = datastore.SetConfig(datastore.ConfGenresUpdated, strconv.Itoa(int(time.Now().Unix())))
	if err != nil {
		logger.Error(err)
	}

	logger.Info("Genres updated")
}

type adminGenreCount struct {
	Count int
	Genre steam.AppDetailsGenre
}

func adminQueues(r *http.Request) {

	if val := r.PostForm.Get("change-id"); val != "" {

		logger.Info("Change: " + val)
		appID, _ := strconv.Atoi(val)
		bytes, _ := json.Marshal(queue.AppMessage{
			AppID: appID,
		})
		queue.Produce(queue.AppQueue, bytes)
	}

	if val := r.PostForm.Get("player-id"); val != "" {

		logger.Info("Player: " + val)
		playerID, _ := strconv.Atoi(val)
		bytes, _ := json.Marshal(queue.PlayerMessage{
			PlayerID: playerID,
		})
		queue.Produce(queue.PlayerQueue, bytes)
	}

	if val := r.PostForm.Get("app-id"); val != "" {

		logger.Info("App: " + val)
		appID, _ := strconv.Atoi(val)
		bytes, _ := json.Marshal(queue.AppMessage{
			AppID: appID,
		})
		queue.Produce(queue.AppQueue, bytes)
	}

	if val := r.PostForm.Get("package-id"); val != "" {

		logger.Info("Package: " + val)
		packageID, _ := strconv.Atoi(val)
		bytes, _ := json.Marshal(queue.PackageMessage{
			PackageID: packageID,
		})
		queue.Produce(queue.PackageQueue, bytes)
	}
}

// todo, handle tags that no longer have any games.
func adminTags() {

	// Get tags names
	response, err := http.Get("http://store.steampowered.com/tagdata/populartags/english")
	if err != nil {
		logger.Error(err)
		return
	}
	defer response.Body.Close()

	// Convert to bytes
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logger.Error(err)
		return
	}

	// Unmarshal JSON
	var resp []steamTag
	if err := json.Unmarshal(contents, &resp); err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			logger.Info(err.Error() + " - " + string(contents))
		} else {
			logger.Error(err)
		}
		return
	}

	// Make tag map
	steamTagMap := make(map[int]string)
	for _, v := range resp {
		steamTagMap[v.Tagid] = v.Name
	}

	// Get apps from mysql
	filter := url.Values{}
	filter.Set("tags_depth", "2")

	apps, err := mysql.SearchApps(filter, 0, "", []string{"name", "price_final", "price_discount", "tags"})
	if err != nil {
		logger.Error(err)
	}

	counts := make(map[int]*adminTag)
	for _, app := range apps {

		tags, err := app.GetTags()
		if err != nil {
			logger.Error(err)
			continue
		}

		for _, tag := range tags {

			if _, ok := counts[tag]; ok {
				counts[tag].count++
				counts[tag].totalPrice = counts[tag].totalPrice + app.PriceFinal
				counts[tag].totalDiscount = counts[tag].totalDiscount + app.PriceDiscount
				counts[tag].name = steamTagMap[tag]
			} else {
				counts[tag] = &adminTag{
					name:          steamTagMap[tag],
					count:         1,
					totalPrice:    app.PriceFinal,
					totalDiscount: app.PriceDiscount,
				}
			}
		}
	}

	var wg sync.WaitGroup

	for k, v := range counts {

		wg.Add(1)
		go func(k int, v *adminTag) {

			err := mysql.SaveOrUpdateTag(k, mysql.Tag{
				Apps:         v.count,
				MeanPrice:    v.GetMeanPrice(),
				MeanDiscount: v.GetMeanDiscount(),
				Name:         v.name,
			})
			if err != nil {
				logger.Error(err)
			}

			wg.Done()

		}(k, v)
	}
	wg.Wait()

	err = datastore.SetConfig(datastore.ConfTagsUpdated, strconv.Itoa(int(time.Now().Unix())))
	if err != nil {
		logger.Error(err)
	}

	logger.Info("Tags updated")
}

type adminTag struct {
	name          string
	count         int
	totalPrice    int
	totalDiscount int
}

func (t adminTag) GetMeanPrice() float64 {
	return float64(t.totalPrice) / float64(t.count)
}

func (t adminTag) GetMeanDiscount() float64 {
	return float64(t.totalDiscount) / float64(t.count)
}

type steamTag struct {
	Tagid int    `json:"tagid"`
	Name  string `json:"name"`
}

func adminRanks() {

	logger.Info("Ranks updated started")

	playersToRank := 1000
	timeStart := time.Now().Unix()

	oldKeys, err := datastore.GetRankKeys()
	if err != nil {
		logger.Error(err)
		return
	}

	newRanks := make(map[int]*datastore.Rank)
	var players []*datastore.Player

	var wg sync.WaitGroup

	for _, v := range []string{"-level", "-games_count", "-badges_count", "-play_time", "-friends_count"} {

		wg.Add(1)
		go func(column string) {

			players, err = datastore.GetPlayers(column, playersToRank)
			if err != nil {
				logger.Error(err)
				return
			}

			for _, v := range players {
				newRanks[v.PlayerID] = datastore.NewRankFromPlayer(*v)
				delete(oldKeys, v.PlayerID)
			}

			wg.Done()
		}(v)

	}
	wg.Wait()

	// Convert new ranks to slice
	var ranks []*datastore.Rank
	for _, v := range newRanks {
		ranks = append(ranks, v)
	}

	// Make ranks
	var prev int
	var rank int

	rank = 0
	sort.Slice(ranks, func(i, j int) bool { return ranks[i].Level > ranks[j].Level })
	for _, v := range ranks {
		if v.Level != prev {
			rank++
		}
		v.UpdatedAt = time.Now()
		v.LevelRank = rank
		prev = v.Level
	}

	rank = 0
	sort.Slice(ranks, func(i, j int) bool { return ranks[i].GamesCount > ranks[j].GamesCount })
	for _, v := range ranks {
		if v.GamesCount != prev {
			rank++
		}
		v.UpdatedAt = time.Now()
		v.GamesRank = rank
		prev = v.GamesCount
	}

	rank = 0
	sort.Slice(ranks, func(i, j int) bool { return ranks[i].BadgesCount > ranks[j].BadgesCount })
	for _, v := range ranks {
		if v.BadgesCount != prev {
			rank++
		}
		v.UpdatedAt = time.Now()
		v.BadgesRank = rank
		prev = v.BadgesCount
	}

	rank = 0
	sort.Slice(ranks, func(i, j int) bool { return ranks[i].PlayTime > ranks[j].PlayTime })
	for _, v := range ranks {
		if v.PlayTime != prev {
			rank++
		}
		v.UpdatedAt = time.Now()
		v.PlayTimeRank = rank
		prev = v.PlayTime
	}

	rank = 0
	sort.Slice(ranks, func(i, j int) bool { return ranks[i].FriendsCount > ranks[j].FriendsCount })
	for _, v := range ranks {
		if v.FriendsCount != prev {
			rank++
		}
		v.UpdatedAt = time.Now()
		v.FriendsRank = rank
		prev = v.FriendsCount
	}

	// Update ranks
	err = datastore.BulkSaveRanks(ranks)
	if err != nil {
		logger.Error(err)
		return
	}

	// Remove old ranks
	err = datastore.BulkDeleteRanks(oldKeys)
	if err != nil {
		logger.Error(err)
		return
	}

	//
	err = datastore.SetConfig(datastore.ConfRanksUpdated, strconv.Itoa(int(time.Now().Unix())))
	if err != nil {
		logger.Error(err)
	}

	logger.Info("Ranks updated in " + strconv.FormatInt(time.Now().Unix()-timeStart, 10) + " seconds")
}
