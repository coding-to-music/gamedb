package web

import (
	"encoding/json"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/go-chi/chi"
	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/memcache"
	"github.com/steam-authority/steam-authority/mysql"
	"github.com/steam-authority/steam-authority/queue"
	"github.com/steam-authority/steam-authority/steami"
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
	case "count-developers":
		go adminDevelopers()
	case "wipe-memcache":
		go adminMemcache()
	case "count-publishers":
		go adminPublishers()
	case "disable-consumers":
		go adminDisableConsumers()
	}

	// Redirect away after action
	if option != "" {
		http.Redirect(w, r, "/admin?"+option, 302)
		return
	}

	// Get configs for times
	configs, err := mysql.GetConfigs([]string{
		mysql.ConfTagsUpdated,
		mysql.ConfGenresUpdated,
		mysql.ConfGenresUpdated,
		mysql.ConfDonationsUpdated,
		mysql.ConfRanksUpdated,
		mysql.ConfAddedAllApps,
		mysql.ConfDevelopersUpdated,
		mysql.ConfPublishersUpdated,
	})
	logger.Error(err)

	// Template
	t := adminTemplate{}
	t.Fill(w, r, "Admin")
	t.Configs = configs

	returnTemplate(w, r, "admin", t)
	return
}

type adminTemplate struct {
	GlobalTemplate
	Errors  []string
	Configs map[string]mysql.Config
}

func adminDisableConsumers() {

}

func adminApps() {

	// Get apps
	// todo, page through results
	apps, _, err := steami.Steam().GetAppList(steam.GetAppList{})
	if err != nil {
		logger.Error(err)
		return
	}

	for _, v := range apps.Apps {
		bytes, _ := json.Marshal(queue.RabbitMessageApp{
			AppID:    v.AppID,
			ChangeID: 0,
			Time:     time.Now(),
		})

		queue.Produce(queue.AppQueue, bytes)
	}

	//
	err = mysql.SetConfig(mysql.ConfAddedAllApps, strconv.Itoa(int(time.Now().Unix())))
	logger.Error(err)

	logger.Info(strconv.Itoa(len(apps.Apps)) + " apps added to rabbit")
}

func adminDeploy() {

	//
	err := mysql.SetConfig(mysql.ConfDeployed, strconv.Itoa(int(time.Now().Unix())))
	logger.Error(err)
}

func adminDonations() {

	donations, err := datastore.GetDonations(0, 0)
	if err != nil {
		logger.Error(err)
		return
	}

	// map[player]total
	counts := make(map[int64]int)

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
	err = mysql.SetConfig(mysql.ConfDonationsUpdated, strconv.Itoa(int(time.Now().Unix())))
	logger.Error(err)

	logger.Info("Updated " + strconv.Itoa(len(counts)) + " player donation counts")
}

func adminGenres() {

	// Get current genres, to delete old ones
	genres, err := mysql.GetAllGenres()
	if err != nil {
		logger.Error(err)
		return
	}

	genresToDelete := map[int]int{}
	for _, v := range genres {
		genresToDelete[v.ID] = v.ID
	}

	// Get apps
	filter := url.Values{}
	filter.Set("genres_depth", "3")

	apps, err := mysql.SearchApps(filter, 0, "", []string{})
	logger.Error(err)

	counts := make(map[int]*adminGenreCount)

	for _, app := range apps {

		genres, err := app.GetGenres()
		if err != nil {
			logger.Error(err)
			continue
		}

		for _, genre := range genres {

			delete(genresToDelete, genre.ID)

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

	// Delete old publishers
	for _, v := range genresToDelete {

		wg.Add(1)
		go func() {

			err := mysql.DeleteGenre(v)
			logger.Error(err)

			wg.Done()
		}()
	}

	// Update current publishers
	for _, v := range counts {

		wg.Add(1)
		go func(v *adminGenreCount) {

			err := mysql.SaveOrUpdateGenre(v.Genre.ID, v.Genre.Description, v.Count)
			logger.Error(err)

			wg.Done()

		}(v)
	}
	wg.Wait()

	//
	err = mysql.SetConfig(mysql.ConfGenresUpdated, strconv.Itoa(int(time.Now().Unix())))
	logger.Error(err)

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
		bytes, _ := json.Marshal(queue.RabbitMessageApp{
			AppID: appID,
			Time:  time.Now(),
		})
		queue.Produce(queue.AppQueue, bytes)
	}

	if val := r.PostForm.Get("player-id"); val != "" {

		logger.Info("Player: " + val)
		playerID, _ := strconv.ParseInt(val, 10, 64)
		bytes, _ := json.Marshal(queue.RabbitMessagePlayer{
			PlayerID: playerID,
			Time:     time.Now(),
		})
		queue.Produce(queue.PlayerQueue, bytes)
	}

	if val := r.PostForm.Get("app-id"); val != "" {

		logger.Info("App: " + val)
		appID, _ := strconv.Atoi(val)
		bytes, _ := json.Marshal(queue.RabbitMessageApp{
			AppID: appID,
			Time:  time.Now(),
		})
		queue.Produce(queue.AppQueue, bytes)
	}

	if val := r.PostForm.Get("package-id"); val != "" {

		logger.Info("Package: " + val)
		packageID, _ := strconv.Atoi(val)
		bytes, _ := json.Marshal(queue.PackageMessage{
			PackageID: packageID,
			Time:      time.Now(),
		})
		queue.Produce(queue.PackageQueue, bytes)
	}
}

func adminPublishers() {

	// Get current publishers, to delete old ones
	publishers, err := mysql.GetAllPublishers()
	if err != nil {
		logger.Error(err)
		return
	}

	pubsToDelete := map[string]int{}
	for _, v := range publishers {
		pubsToDelete[v.Name] = v.ID
	}

	// Get apps from mysql
	apps, err := mysql.SearchApps(url.Values{}, 0, "", []string{"name", "price_final", "price_discount", "publishers"})
	logger.Error(err)

	counts := make(map[string]*adminDeveloper)

	for _, app := range apps {

		publishers, err := app.GetPublishers()
		if err != nil {
			logger.Error(err)
			continue
		}

		if len(publishers) == 0 {
			publishers = []string{"No Publisher"}
		}

		for _, key := range publishers {

			key = strings.ToLower(key)

			delete(pubsToDelete, key)

			if _, ok := counts[key]; ok {
				counts[key].count++
				counts[key].totalPrice = counts[key].totalPrice + app.PriceFinal
				counts[key].totalDiscount = counts[key].totalDiscount + app.PriceDiscount
			} else {
				counts[key] = &adminDeveloper{
					count:         1,
					totalPrice:    app.PriceFinal,
					totalDiscount: app.PriceDiscount,
					name:          app.GetName(),
				}
			}
		}
	}

	var wg sync.WaitGroup

	// Delete old publishers
	for _, v := range pubsToDelete {

		wg.Add(1)
		go func() {

			err := mysql.DeletePublisher(v)
			logger.Error(err)

			wg.Done()
		}()
	}

	// Update current publishers
	for k, v := range counts {

		wg.Add(1)
		go func(k string, v *adminDeveloper) {

			err := mysql.SaveOrUpdatePublisher(k, mysql.Publisher{
				Apps:         v.count,
				MeanPrice:    v.GetMeanPrice(),
				MeanDiscount: v.GetMeanDiscount(),
				Name:         v.name,
			})
			logger.Error(err)

			wg.Done()

		}(k, v)
	}

	wg.Wait()

	err = mysql.SetConfig(mysql.ConfPublishersUpdated, strconv.Itoa(int(time.Now().Unix())))
	logger.Error(err)

	logger.Info("Publishers updated")
}

func adminDevelopers() {

	// Get current publishers, to delete old ones
	developers, err := mysql.GetAllPublishers()
	if err != nil {
		logger.Error(err)
		return
	}

	devsToDelete := map[string]int{}
	for _, v := range developers {
		devsToDelete[v.Name] = v.ID
	}

	// Get apps from mysql
	apps, err := mysql.SearchApps(url.Values{}, 0, "", []string{"name", "price_final", "price_discount", "developers"})
	logger.Error(err)

	counts := make(map[string]*adminDeveloper)

	for _, app := range apps {

		developers, err := app.GetDevelopers()
		if err != nil {
			logger.Error(err)
			continue
		}

		if len(developers) == 0 {
			developers = []string{"No Developer"}
		}

		for _, key := range developers {

			key = strings.ToLower(key)

			delete(devsToDelete, key)

			if _, ok := counts[key]; ok {
				counts[key].count++
				counts[key].totalPrice = counts[key].totalPrice + app.PriceFinal
				counts[key].totalDiscount = counts[key].totalDiscount + app.PriceDiscount
			} else {
				counts[key] = &adminDeveloper{
					count:         1,
					totalPrice:    app.PriceFinal,
					totalDiscount: app.PriceDiscount,
					name:          app.GetName(),
				}
			}
		}
	}

	var wg sync.WaitGroup

	// Delete old developers
	for _, v := range devsToDelete {

		wg.Add(1)
		go func() {

			mysql.DeleteDeveloper(v)

			wg.Done()
		}()
	}

	// Update current developers
	for k, v := range counts {

		wg.Add(1)
		go func(k string, v *adminDeveloper) {

			err := mysql.SaveOrUpdateDeveloper(k, mysql.Developer{
				Apps:         v.count,
				MeanPrice:    v.GetMeanPrice(),
				MeanDiscount: v.GetMeanDiscount(),
				Name:         v.name,
			})
			logger.Error(err)

			wg.Done()

		}(k, v)
	}
	wg.Wait()

	err = mysql.SetConfig(mysql.ConfDevelopersUpdated, strconv.Itoa(int(time.Now().Unix())))
	logger.Error(err)

	logger.Info("Developers updated")
}

type adminDeveloper struct {
	name          string
	count         int
	totalPrice    int
	totalDiscount int
}

func (t adminDeveloper) GetMeanPrice() float64 {
	return float64(t.totalPrice) / float64(t.count)
}

func (t adminDeveloper) GetMeanDiscount() float64 {
	return float64(t.totalDiscount) / float64(t.count)
}

func adminTags() {

	// Get current tags, to delete old ones
	tags, err := mysql.GetAllTags()
	if err != nil {
		logger.Error(err)
		return
	}

	tagsToDelete := map[int]int{}
	for _, tag := range tags {
		tagsToDelete[tag.ID] = tag.ID
	}

	// Get tag names from Steam
	tagsResp, _, err := steami.Steam().GetTags()
	logger.Error(err)

	steamTagMap := make(map[int]string)
	for _, v := range tagsResp.Tags {
		steamTagMap[v.TagID] = v.Name
	}

	// Get apps from mysql
	filter := url.Values{}
	filter.Set("tags_depth", "2")

	apps, err := mysql.SearchApps(filter, 0, "", []string{"name", "price_final", "price_discount", "tags"})
	logger.Error(err)

	counts := make(map[int]*adminTag)
	for _, app := range apps {

		tags, err := app.GetTags()
		if err != nil {
			logger.Error(err)
			continue
		}

		for _, key := range tags {

			delete(tagsToDelete, key)

			if _, ok := counts[key]; ok {
				counts[key].count++
				counts[key].totalPrice = counts[key].totalPrice + app.PriceFinal
				counts[key].totalDiscount = counts[key].totalDiscount + app.PriceDiscount
				counts[key].name = steamTagMap[key]
			} else {
				counts[key] = &adminTag{
					name:          steamTagMap[key],
					count:         1,
					totalPrice:    app.PriceFinal,
					totalDiscount: app.PriceDiscount,
				}
			}
		}
	}

	var wg sync.WaitGroup

	// Delete old tags
	for _, v := range tagsToDelete {

		wg.Add(1)
		go func() {

			err := mysql.DeleteTag(v)
			logger.Error(err)

			wg.Done()
		}()
	}

	// Update current tags
	for k, v := range counts {

		wg.Add(1)
		go func(k int, v *adminTag) {

			err := mysql.SaveOrUpdateTag(k, mysql.Tag{
				Apps:         v.count,
				MeanPrice:    v.GetMeanPrice(),
				MeanDiscount: v.GetMeanDiscount(),
				Name:         v.name,
			})
			logger.Error(err)

			wg.Done()
		}(k, v)
	}
	wg.Wait()

	err = mysql.SetConfig(mysql.ConfTagsUpdated, strconv.Itoa(int(time.Now().Unix())))
	logger.Error(err)

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

func adminRanks() {

	logger.Info("Ranks updated started")

	playersToRank := 1000
	timeStart := time.Now().Unix()

	oldKeys, err := datastore.GetRankKeys()
	if err != nil {
		logger.Error(err)
		return
	}

	newRanks := make(map[int64]*datastore.Rank)
	var players []datastore.Player

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
				newRanks[v.PlayerID] = datastore.NewRankFromPlayer(v)
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
	sort.Slice(ranks, func(i, j int) bool {
		return ranks[i].Level > ranks[j].Level
	})
	for _, v := range ranks {
		if v.Level != prev {
			rank++
		}
		v.UpdatedAt = time.Now()
		v.LevelRank = rank
		prev = v.Level
	}

	rank = 0
	sort.Slice(ranks, func(i, j int) bool {
		return ranks[i].GamesCount > ranks[j].GamesCount
	})
	for _, v := range ranks {
		if v.GamesCount != prev {
			rank++
		}
		v.UpdatedAt = time.Now()
		v.GamesRank = rank
		prev = v.GamesCount
	}

	rank = 0
	sort.Slice(ranks, func(i, j int) bool {
		return ranks[i].BadgesCount > ranks[j].BadgesCount
	})
	for _, v := range ranks {
		if v.BadgesCount != prev {
			rank++
		}
		v.UpdatedAt = time.Now()
		v.BadgesRank = rank
		prev = v.BadgesCount
	}

	rank = 0
	sort.Slice(ranks, func(i, j int) bool {
		return ranks[i].PlayTime > ranks[j].PlayTime
	})
	for _, v := range ranks {
		if v.PlayTime != prev {
			rank++
		}
		v.UpdatedAt = time.Now()
		v.PlayTimeRank = rank
		prev = v.PlayTime
	}

	rank = 0
	sort.Slice(ranks, func(i, j int) bool {
		return ranks[i].FriendsCount > ranks[j].FriendsCount
	})
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
	err = mysql.SetConfig(mysql.ConfRanksUpdated, strconv.Itoa(int(time.Now().Unix())))
	logger.Error(err)

	logger.Info("Ranks updated in " + strconv.FormatInt(time.Now().Unix()-timeStart, 10) + " seconds")
}

func adminMemcache() {

	err := memcache.Wipe()
	logger.Error(err)

	logger.Info("Memcache wiped")
}
