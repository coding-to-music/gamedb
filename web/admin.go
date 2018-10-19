package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/Jleagle/steam-go/steam"
	"github.com/go-chi/chi"
	"github.com/gosimple/slug"
	"github.com/steam-authority/steam-authority/db"
	"github.com/steam-authority/steam-authority/helpers"
	"github.com/steam-authority/steam-authority/logging"
	"github.com/steam-authority/steam-authority/memcache"
	"github.com/steam-authority/steam-authority/queue"
)

func AdminHandler(w http.ResponseWriter, r *http.Request) {

	option := chi.URLParam(r, "option")

	switch option {
	case "re-add-all-apps":
		go adminApps()
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
	case "dev":
		go adminDev()
	}

	// Redirect away after action
	if option != "" {
		http.Redirect(w, r, "/admin?"+option, 302)
		return
	}

	// Get configs for times
	configs, err := db.GetConfigs([]string{
		db.ConfTagsUpdated,
		db.ConfGenresUpdated,
		db.ConfGenresUpdated,
		db.ConfDonationsUpdated,
		db.ConfRanksUpdated,
		db.ConfAddedAllApps,
		db.ConfDevelopersUpdated,
		db.ConfPublishersUpdated,
	})
	logging.Error(err)

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
	Configs map[string]db.Config
}

func adminDisableConsumers() {

}

func adminApps() {

	// Get apps
	// todo, page through results
	apps, _, err := helpers.GetSteam().GetAppList(steam.GetAppListOptions{})
	if err != nil {
		logging.Error(err)
		return
	}

	for _, v := range apps.Apps {
		queue.Produce(queue.QueueApps, []byte(strconv.Itoa(v.AppID)))
	}

	//
	err = db.SetConfig(db.ConfAddedAllApps, strconv.Itoa(int(time.Now().Unix())))
	logging.Error(err)

	logging.Info(strconv.Itoa(len(apps.Apps)) + " apps added to rabbit")
}

func adminDonations() {

	donations, err := db.GetDonations(0, 0)
	if err != nil {
		logging.Error(err)
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
		player, err := db.GetPlayer(k)
		if err != nil {
			logging.Error(err)
			continue
		}

		player.Donated = v
		_, err = db.SaveKind(player.GetKey(), player)
	}

	//
	err = db.SetConfig(db.ConfDonationsUpdated, strconv.Itoa(int(time.Now().Unix())))
	logging.Error(err)

	logging.Info("Updated " + strconv.Itoa(len(counts)) + " player donation counts")
}

func adminQueues(r *http.Request) {

	if val := r.PostForm.Get("change-id"); val != "" {

		logging.Info("Change: " + val)
		//queue.Produce(queue.ProduceOptions{queue.chang, []byte(val)), 1})
	}

	if val := r.PostForm.Get("player-id"); val != "" {

		logging.Info("Player: " + val)
		playerID, err := strconv.ParseInt(val, 10, 64)
		logging.Error(err)
		bytes, _ := json.Marshal(queue.RabbitMessageProfile{
			PlayerID: playerID,
			Time:     time.Now(),
		})
		queue.Produce(queue.QueueProfiles, bytes)
	}

	if val := r.PostForm.Get("app-id"); val != "" {

		logging.Info("App: " + val)
		queue.Produce(queue.QueueApps, []byte(val))
	}

	if val := r.PostForm.Get("package-id"); val != "" {

		logging.Info("Package: " + val)
		queue.Produce(queue.QueuePackages, []byte(val))
	}
}

func adminGenres() {

	// Get current genres, to delete old ones
	genres, err := db.GetAllGenres()
	if err != nil {
		logging.Error(err)
		return
	}

	genresToDelete := map[int]int{}
	for _, v := range genres {
		genresToDelete[v.ID] = v.ID
	}

	// Get apps
	filter := url.Values{}
	filter.Set("genres_depth", "3")

	apps, err := db.SearchApps(filter, 0, 0, "", []string{})
	logging.Error(err)

	counts := make(map[int]*adminGenreCount)

	for _, app := range apps {

		genres, err := app.GetGenres()
		if err != nil {
			logging.Error(err)
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

			err := db.DeleteGenre(v)
			logging.Error(err)

			wg.Done()
		}()
	}

	// Update current publishers
	for _, v := range counts {

		wg.Add(1)
		go func(v *adminGenreCount) {

			err := db.SaveOrUpdateGenre(v.Genre.ID, v.Genre.Description, v.Count)
			logging.Error(err)

			wg.Done()

		}(v)
	}
	wg.Wait()

	//
	err = db.SetConfig(db.ConfGenresUpdated, strconv.Itoa(int(time.Now().Unix())))
	logging.Error(err)

	logging.Info("Genres updated")
}

type adminGenreCount struct {
	Count int
	Genre steam.AppDetailsGenre
}

func adminPublishers() {

	fmt.Println("Publishers Updating")

	// Get current publishers, to delete old ones
	currentPublishers, err := db.GetAllPublishers()
	if err != nil {
		logging.Error(err)
		return
	}

	pubsToDelete := map[string]int{}
	for _, publisherRow := range currentPublishers {
		pubsToDelete[slug.Make(publisherRow.Name)] = publisherRow.ID
	}

	// Get apps from mysql
	apps, err := db.SearchApps(url.Values{}, 0, 1, "", []string{"name", "price_final", "publishers", "reviews_score"})
	if err != nil {
		logging.Error(err)
		return
	}

	publishersToAdd := make(map[string]*adminPublisher)

	for _, app := range apps {

		appPublishers, err := app.GetPublishers()
		if err != nil {
			logging.Error(err)
			continue
		}

		if len(appPublishers) == 0 {
			appPublishers = []string{""}
		}

		for _, publisher := range appPublishers {

			key := slug.Make(publisher)

			delete(pubsToDelete, key)

			if _, ok := publishersToAdd[key]; ok {
				publishersToAdd[key].count++
				publishersToAdd[key].totalPrice += app.PriceFinal
				publishersToAdd[key].totalScore += app.ReviewsScore
			} else {
				publishersToAdd[key] = &adminPublisher{
					count:      1,
					totalPrice: app.PriceFinal,
					totalScore: app.ReviewsScore,
					name:       publisher,
				}
			}
		}
	}

	var wg sync.WaitGroup

	// Delete old publishers
	wg.Add(1)
	go func() {

		var pubsToDeleteSlice []int
		for _, v := range pubsToDelete {
			pubsToDeleteSlice = append(pubsToDeleteSlice, v)
		}

		err := db.DeletePublishers(pubsToDeleteSlice)
		logging.Error(err)

		wg.Done()
	}()

	// Update current publishers
	var limit int
	for k, v := range publishersToAdd {

		if limit >= 5 {
			wg.Wait()
		}

		limit++
		wg.Add(1)
		go func(k string, v *adminPublisher) {

			err := db.SaveOrUpdatePublisher(k, db.Publisher{
				Apps:      v.count,
				MeanPrice: v.GetMeanPrice(),
				MeanScore: v.GetMeanScore(),
				Name:      v.name,
			})
			logging.Error(err)

			limit--
			wg.Done()

		}(k, v)
	}

	wg.Wait()

	err = db.SetConfig(db.ConfPublishersUpdated, strconv.Itoa(int(time.Now().Unix())))
	logging.Error(err)

	logging.Info("Publishers updated")
}

type adminPublisher struct {
	name       string
	count      int
	totalPrice int
	totalScore float64
}

func (p adminPublisher) GetMeanPrice() float64 {
	return float64(p.totalPrice) / float64(p.count)
}

func (p adminPublisher) GetMeanScore() float64 {
	return float64(p.totalScore) / float64(p.count)
}

func adminDevelopers() {

	// Get current publishers, to delete old ones
	developers, err := db.GetAllPublishers()
	if err != nil {
		logging.Error(err)
		return
	}

	devsToDelete := map[string]int{}
	for _, v := range developers {
		devsToDelete[v.Name] = v.ID
	}

	// Get apps from mysql
	apps, err := db.SearchApps(url.Values{}, 0, 1, "", []string{"name", "price_final", "developers"})
	logging.Error(err)

	counts := make(map[string]*adminDeveloper)

	for _, app := range apps {

		developers, err := app.GetDevelopers()
		if err != nil {
			logging.Error(err)
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
				counts[key].totalPrice += app.PriceFinal
				counts[key].totalScore += app.ReviewsScore
			} else {
				counts[key] = &adminDeveloper{
					count:      1,
					totalPrice: app.PriceFinal,
					totalScore: app.ReviewsScore,
					name:       app.GetName(),
				}
			}
		}
	}

	var wg sync.WaitGroup

	// Delete old developers
	for _, v := range devsToDelete {

		wg.Add(1)
		go func() {

			db.DeleteDeveloper(v)

			wg.Done()
		}()
	}

	// Update current developers
	var limit int
	for k, v := range counts {

		if limit >= 5 {
			wg.Wait()
		}

		limit++
		wg.Add(1)
		go func(k string, v *adminDeveloper) {

			err := db.SaveOrUpdateDeveloper(k, db.Developer{
				Apps:      v.count,
				MeanPrice: v.GetMeanPrice(),
				MeanScore: v.GetMeanScore(),
				Name:      v.name,
			})
			logging.Error(err)

			limit--
			wg.Done()

		}(k, v)
	}
	wg.Wait()

	err = db.SetConfig(db.ConfDevelopersUpdated, strconv.Itoa(int(time.Now().Unix())))
	logging.Error(err)

	logging.Info("Developers updated")
}

type adminDeveloper struct {
	name       string
	count      int
	totalPrice int
	totalScore float64
}

func (t adminDeveloper) GetMeanPrice() float64 {
	return float64(t.totalPrice) / float64(t.count)
}

func (t adminDeveloper) GetMeanScore() float64 {
	return float64(t.totalScore) / float64(t.count)
}

func adminTags() {

	// Get current tags, to delete old ones
	tags, err := db.GetAllTags()
	if err != nil {
		logging.Error(err)
		return
	}

	tagsToDelete := map[int]int{}
	for _, tag := range tags {
		tagsToDelete[tag.ID] = tag.ID
	}

	// Get tag names from Steam
	tagsResp, _, err := helpers.GetSteam().GetTags()
	logging.Error(err)

	steamTagMap := make(map[int]string)
	for _, v := range tagsResp.Tags {
		steamTagMap[v.TagID] = v.Name
	}

	// Get apps from mysql
	filter := url.Values{}
	filter.Set("tags_depth", "2")

	apps, err := db.SearchApps(filter, 0, 1, "", []string{"name", "price_final", "tags"})
	logging.Error(err)

	counts := make(map[int]*adminTag)
	for _, app := range apps {

		tags, err := app.GetTagIDs()
		if err != nil {
			logging.Error(err)
			continue
		}

		for _, key := range tags {

			delete(tagsToDelete, key)

			if _, ok := counts[key]; ok {
				counts[key].count++
				counts[key].totalPrice += app.PriceFinal
				counts[key].totalScore += app.ReviewsScore
				counts[key].name = steamTagMap[key]
			} else {
				counts[key] = &adminTag{
					name:       steamTagMap[key],
					count:      1,
					totalPrice: app.PriceFinal,
					totalScore: app.ReviewsScore,
				}
			}
		}
	}

	var wg sync.WaitGroup

	// Delete old tags
	for _, v := range tagsToDelete {

		wg.Add(1)
		go func() {

			err := db.DeleteTag(v)
			logging.Error(err)

			wg.Done()
		}()
	}

	// Update current tags
	var limit int
	for k, v := range counts {

		if limit >= 5 {
			wg.Wait()
		}

		limit++
		wg.Add(1)
		go func(k int, v *adminTag) {

			err := db.SaveOrUpdateTag(k, db.Tag{
				Apps:      v.count,
				MeanPrice: v.GetMeanPrice(),
				MeanScore: v.GetMeanScore(),
				Name:      v.name,
			})
			logging.Error(err)

			limit--
			wg.Done()
		}(k, v)
	}
	wg.Wait()

	err = db.SetConfig(db.ConfTagsUpdated, strconv.Itoa(int(time.Now().Unix())))
	logging.Error(err)

	logging.Info("Tags updated")
}

type adminTag struct {
	name       string
	count      int
	totalPrice int
	totalScore float64
}

func (t adminTag) GetMeanPrice() float64 {
	return float64(t.totalPrice) / float64(t.count)
}

func (t adminTag) GetMeanScore() float64 {
	return float64(t.totalScore) / float64(t.count)
}

func adminRanks() {

	logging.Info("Ranks updated started")

	timeStart := time.Now().Unix()

	oldKeys, err := db.GetRankKeys()
	if err != nil {
		logging.Error(err)
		return
	}

	newRanks := make(map[int64]*db.PlayerRank)
	var players []db.Player

	var wg sync.WaitGroup

	for _, v := range []string{"-level", "-games_count", "-badges_count", "-play_time", "-friends_count"} {

		wg.Add(1)
		go func(column string) {

			players, err = db.GetAllPlayers(column, db.PlayersToRank)
			if err != nil {
				logging.Error(err)
				return
			}

			for _, v := range players {
				newRanks[v.PlayerID] = db.NewRankFromPlayer(v)
				delete(oldKeys, v.PlayerID)
			}

			wg.Done()
		}(v)

	}
	wg.Wait()

	// Convert new ranks to slice
	var ranks []*db.PlayerRank
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
		return ranks[i].Games > ranks[j].Games
	})
	for _, v := range ranks {
		if v.Games != prev {
			rank++
		}
		v.UpdatedAt = time.Now()
		v.GamesRank = rank
		prev = v.Games
	}

	rank = 0
	sort.Slice(ranks, func(i, j int) bool {
		return ranks[i].Badges > ranks[j].Badges
	})
	for _, v := range ranks {
		if v.Badges != prev {
			rank++
		}
		v.UpdatedAt = time.Now()
		v.BadgesRank = rank
		prev = v.Badges
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
		return ranks[i].Friends > ranks[j].Friends
	})
	for _, v := range ranks {
		if v.Friends != prev {
			rank++
		}
		v.UpdatedAt = time.Now()
		v.FriendsRank = rank
		prev = v.Friends
	}

	// Make kinds
	var kinds []db.Kind
	for _, v := range ranks {
		kinds = append(kinds, *v)
	}

	// Update ranks
	err = db.BulkSaveKinds(kinds, db.KindPlayerRank, false)
	if err != nil {
		logging.Error(err)
		return
	}

	// Remove old ranks
	var keysToDelete []*datastore.Key
	for _, v := range oldKeys {
		keysToDelete = append(keysToDelete, v)
	}

	err = db.BulkDeleteKinds(keysToDelete, false)
	if err != nil {
		logging.Error(err)
		return
	}

	// Update config
	err = db.SetConfig(db.ConfRanksUpdated, strconv.Itoa(int(time.Now().Unix())))
	logging.Error(err)

	logging.Info("Ranks updated in " + strconv.FormatInt(time.Now().Unix()-timeStart, 10) + " seconds")
}

func adminMemcache() {

	err := memcache.Wipe()
	logging.Error(err)

	logging.Info("Memcache wiped")
}

func adminDev() {

	//return
	//
	//logging.Info("Dev")
	//
	//players, err := db.GetAllPlayers("__key__", 0)
	//
	//logging.Info("Got players")
	//
	//if err != nil {
	//
	//	logging.Error(err)
	//
	//	if _, ok := err.(*ds.ErrFieldMismatch); ok {
	//
	//	} else {
	//		return
	//	}
	//}
	//
	//for _, v := range players {
	//	//v.Games = ""
	//	err := v.Save()
	//	logging.Error(err)
	//	fmt.Print(".")
	//}
	//
	//logging.Info("Done")
}
