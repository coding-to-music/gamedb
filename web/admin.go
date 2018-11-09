package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/logging"
	"github.com/gamedb/website/memcache"
	"github.com/gamedb/website/queue"
	"github.com/go-chi/chi"
	"github.com/gosimple/slug"
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
		logging.Error(err)
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

		message := queue.RabbitMessageProfile{}
		message.Fill(r, playerID, db.PlayerUpdateAdmin)

		queue.Produce(queue.QueueProfiles, message.ToBytes())
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

	logging.InfoL("Genres updating")

	gorm, err := db.GetMySQLClient()
	if err != nil {
		logging.Error(err)
		return
	}

	// Get current genres, to delete old ones
	currentGenres, err := db.GetAllGenres()
	if err != nil {
		logging.Error(err)
		return
	}

	genresToDelete := map[int]int{}
	for _, v := range currentGenres {
		genresToDelete[v.ID] = v.ID
	}

	// Get apps from mysql
	appsWithGenres, err := db.GetAppsWithGenres()
	logging.Error(err)

	fmt.Println("Found " + strconv.Itoa(len(appsWithGenres)) + " apps with genres")

	newGenres := make(map[int]*statsRow)
	for _, app := range appsWithGenres {

		appGenres, err := app.GetGenres()
		if err != nil {
			logging.Error(err)
			continue
		}

		if len(appGenres) == 0 {
			appGenres = []steam.AppDetailsGenre{{ID: 0, Description: "No Genre"}}
		}

		for _, genre := range appGenres {

			delete(genresToDelete, genre.ID)

			for code := range steam.Countries {

				price := app.GetPrice(code)

				if _, ok := newGenres[genre.ID]; ok {
					newGenres[genre.ID].count++
					newGenres[genre.ID].totalPrice[code] += price.Final
					newGenres[genre.ID].totalScore[code] += app.ReviewsScore
				} else {
					newGenres[genre.ID] = &statsRow{
						name:       genre.Description,
						count:      1,
						totalPrice: map[steam.CountryCode]int{code: price.Final},
						totalScore: map[steam.CountryCode]float64{code: app.ReviewsScore},
					}
				}
			}
		}
	}

	var limit int
	var wg sync.WaitGroup

	// Delete old genres
	limit++
	wg.Add(1)
	go func() {

		var genresToDeleteSlice []int
		for _, v := range genresToDelete {
			genresToDeleteSlice = append(genresToDeleteSlice, v)
		}

		err := db.DeleteGenres(genresToDeleteSlice)
		logging.Error(err)

		limit--
		wg.Done()
	}()

	// Update current genres
	for k, v := range newGenres {

		if limit >= 5 {
			wg.Wait()
		}

		limit++
		wg.Add(1)
		go func(genreID int, v *statsRow) {

			fmt.Println("Updating genre: " + v.name)

			var genre db.Genre

			gorm.Unscoped().FirstOrInit(&genre, db.Genre{ID: genreID})
			if gorm.Error != nil {
				logging.Error(gorm.Error)
			}

			genre.Name = v.name
			genre.Apps = v.GetCount()
			genre.MeanPrice = v.GetMeanPrice()
			genre.MeanScore = v.GetMeanScore()
			genre.DeletedAt = nil

			gorm.Unscoped().Save(&genre)
			if gorm.Error != nil {
				logging.Error(gorm.Error)
			}

			limit--
			wg.Done()

		}(k, v)
	}
	wg.Wait()

	//
	err = db.SetConfig(db.ConfGenresUpdated, strconv.Itoa(int(time.Now().Unix())))
	logging.Error(err)

	logging.Info("Genres updated")
}

func adminPublishers() {

	logging.InfoL("Publishers updating")

	gorm, err := db.GetMySQLClient()
	if err != nil {
		logging.Error(err)
		return
	}

	// Get current publishers, to delete old ones
	currentPublishers, err := db.GetAllPublishers()
	if err != nil {
		logging.Error(err)
		return
	}

	publishersToDelete := map[string]int{}
	for _, publisherRow := range currentPublishers {
		publishersToDelete[slug.Make(publisherRow.Name)] = publisherRow.ID
	}

	// Get apps from mysql
	appsWithPublishers, err := db.GetAppsWithPublishers()
	logging.Error(err)

	fmt.Println("Found " + strconv.Itoa(len(appsWithPublishers)) + " apps with publishers")

	newPublishers := make(map[string]*statsRow)
	for _, app := range appsWithPublishers {

		appPublishers, err := app.GetPublishers()
		if err != nil {
			logging.Error(err)
			continue
		}

		if len(appPublishers) == 0 {
			appPublishers = []string{""}
		}

		for _, publisher := range appPublishers {

			delete(publishersToDelete, publisher)

			for code := range steam.Countries {

				price := app.GetPrice(code)

				if _, ok := newPublishers[publisher]; ok {
					newPublishers[publisher].count++
					newPublishers[publisher].totalPrice[code] += price.Final
					newPublishers[publisher].totalScore[code] += app.ReviewsScore
				} else {
					newPublishers[publisher] = &statsRow{
						name:       publisher,
						count:      1,
						totalPrice: map[steam.CountryCode]int{code: price.Final},
						totalScore: map[steam.CountryCode]float64{code: app.ReviewsScore},
					}
				}
			}
		}
	}

	var limit int
	var wg sync.WaitGroup

	// Delete old publishers
	limit++
	wg.Add(1)
	go func() {

		var pubsToDeleteSlice []int
		for _, v := range publishersToDelete {
			pubsToDeleteSlice = append(pubsToDeleteSlice, v)
		}

		err := db.DeletePublishers(pubsToDeleteSlice)
		logging.Error(err)

		limit--
		wg.Done()
	}()

	// Update current publishers
	for k, v := range newPublishers {

		if limit >= 5 {
			wg.Wait()
		}

		limit++
		wg.Add(1)
		go func(publisherName string, v *statsRow) {

			fmt.Println("Updating publisher: " + publisherName)

			var publisher db.Developer

			gorm.Unscoped().FirstOrInit(&publisher, db.Developer{Name: publisherName})
			if gorm.Error != nil {
				logging.Error(gorm.Error)
			}

			publisher.Name = v.name
			publisher.Apps = v.GetCount()
			publisher.MeanPrice = v.GetMeanPrice()
			publisher.MeanScore = v.GetMeanScore()
			publisher.DeletedAt = nil

			gorm.Unscoped().Save(&publisher)
			if gorm.Error != nil {
				logging.Error(gorm.Error)
			}

			limit--
			wg.Done()

		}(k, v)
	}

	wg.Wait()

	err = db.SetConfig(db.ConfPublishersUpdated, strconv.Itoa(int(time.Now().Unix())))
	logging.Error(err)

	logging.Info("Publishers updated")
}

func adminDevelopers() {

	logging.InfoL("Developers updating")

	gorm, err := db.GetMySQLClient()
	if err != nil {
		logging.Error(err)
		return
	}

	// Get current developers, to delete old ones
	currentDevelopers, err := db.GetAllPublishers()
	if err != nil {
		logging.Error(err)
		return
	}

	developersToDelete := map[string]int{}
	for _, v := range currentDevelopers {
		developersToDelete[v.Name] = v.ID
	}

	// Get apps from mysql
	appsWithDevelopers, err := db.GetAppsWithDevelopers()
	logging.Error(err)

	fmt.Println("Found " + strconv.Itoa(len(appsWithDevelopers)) + " apps with developers")

	newDevelopers := make(map[string]*statsRow)
	for _, app := range appsWithDevelopers {

		appDevelopers, err := app.GetDevelopers()
		if err != nil {
			logging.Error(err)
			continue
		}

		if len(appDevelopers) == 0 {
			appDevelopers = []string{"No Developer"}
		}

		for _, developer := range appDevelopers {

			delete(developersToDelete, developer)

			for code := range steam.Countries {

				price := app.GetPrice(code)

				if _, ok := newDevelopers[developer]; ok {
					newDevelopers[developer].count++
					newDevelopers[developer].totalPrice[code] += price.Final
					newDevelopers[developer].totalScore[code] += app.ReviewsScore
				} else {
					newDevelopers[developer] = &statsRow{
						name:       developer,
						count:      1,
						totalPrice: map[steam.CountryCode]int{code: price.Final},
						totalScore: map[steam.CountryCode]float64{code: app.ReviewsScore},
					}
				}
			}
		}
	}

	var limit int
	var wg sync.WaitGroup

	// Delete old developers
	limit++
	wg.Add(1)
	go func() {

		var devsToDeleteSlice []int
		for _, v := range developersToDelete {
			devsToDeleteSlice = append(devsToDeleteSlice, v)
		}

		db.DeleteDevelopers(devsToDeleteSlice)

		limit--
		wg.Done()
	}()

	// Update current developers
	for k, v := range newDevelopers {

		if limit >= 5 {
			wg.Wait()
		}

		limit++
		wg.Add(1)
		go func(developerName string, v *statsRow) {

			fmt.Println("Updating developer: " + developerName)

			var developer db.Developer

			gorm.Unscoped().FirstOrInit(&developer, db.Developer{Name: developerName})
			if gorm.Error != nil {
				logging.Error(gorm.Error)
			}

			developer.Name = v.name
			developer.Apps = v.GetCount()
			developer.MeanPrice = v.GetMeanPrice()
			developer.MeanScore = v.GetMeanScore()
			developer.DeletedAt = nil

			gorm.Unscoped().Save(&developer)
			if gorm.Error != nil {
				logging.Error(gorm.Error)
			}

			limit--
			wg.Done()

		}(k, v)
	}
	wg.Wait()

	err = db.SetConfig(db.ConfDevelopersUpdated, strconv.Itoa(int(time.Now().Unix())))
	logging.Error(err)

	logging.Info("Developers updated")
}

func adminTags() {

	gorm, err := db.GetMySQLClient()
	if err != nil {
		logging.Error(err)
		return
	}

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

	// Get tags from Steam
	tagsResp, _, err := helpers.GetSteam().GetTags()
	logging.Error(err)

	steamTagMap := tagsResp.GetMap()

	appsWithTags, err := db.GetAppsWithTags()
	logging.Error(err)

	fmt.Println("Found " + strconv.Itoa(len(appsWithTags)) + " apps with tags")

	newTags := make(map[int]*statsRow)
	for _, app := range appsWithTags {

		appTags, err := app.GetTagIDs()
		if err != nil {
			logging.Error(err)
			continue
		}

		for _, tagID := range appTags {

			delete(tagsToDelete, tagID)

			for code := range steam.Countries {

				price := app.GetPrice(code)

				if _, ok := newTags[tagID]; ok {
					newTags[tagID].count++
					newTags[tagID].totalPrice[code] += price.Final
					newTags[tagID].totalScore[code] += app.ReviewsScore
				} else {
					newTags[tagID] = &statsRow{
						name:       steamTagMap[tagID],
						count:      1,
						totalPrice: map[steam.CountryCode]int{code: price.Final},
						totalScore: map[steam.CountryCode]float64{code: app.ReviewsScore},
					}
				}
			}
		}
	}

	var limit int
	var wg sync.WaitGroup

	// Delete old tags
	limit++
	wg.Add(1)
	go func() {

		var tagsToDeleteSlice []int
		for _, v := range tagsToDelete {
			tagsToDeleteSlice = append(tagsToDeleteSlice, v)
		}

		err := db.DeleteTags(tagsToDeleteSlice)
		logging.Error(err)

		limit--
		wg.Done()
	}()

	// Update current tags
	for k, v := range newTags {

		if limit >= 5 {
			wg.Wait()
		}

		limit++
		wg.Add(1)
		go func(tagID int, v *statsRow) {

			fmt.Println("Updating tag: " + strconv.Itoa(tagID))

			var tag db.Tag

			gorm.Unscoped().FirstOrInit(&tag, db.Tag{ID: tagID})
			if gorm.Error != nil {
				logging.Error(gorm.Error)
			}

			tag.Name = v.name
			tag.Apps = v.GetCount()
			tag.MeanPrice = v.GetMeanPrice()
			tag.MeanScore = v.GetMeanScore()
			tag.DeletedAt = nil

			gorm.Unscoped().Save(&tag)
			if gorm.Error != nil {
				logging.Error(gorm.Error)
			}

			limit--
			wg.Done()

		}(k, v)
	}
	wg.Wait()

	err = db.SetConfig(db.ConfTagsUpdated, strconv.Itoa(int(time.Now().Unix())))
	logging.Error(err)

	logging.Info("Tags updated")
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
	var rank = 0

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

type statsRow struct {
	name       string
	count      int
	totalPrice map[steam.CountryCode]int
	totalScore map[steam.CountryCode]float64
}

func (t statsRow) GetMeanPrice() string {

	means := map[steam.CountryCode]float64{}

	for code, total := range t.totalPrice {
		means[code] = float64(total) / float64(t.GetCount())
	}

	bytes, err := json.Marshal(means)
	logging.Error(err)

	return string(bytes)
}

func (t statsRow) GetMeanScore() string {

	means := map[steam.CountryCode]float64{}

	for code, total := range t.totalScore {
		means[code] = float64(total) / float64(t.GetCount())
	}

	bytes, err := json.Marshal(means)
	logging.Error(err)

	return string(bytes)
}

func (t statsRow) GetCount() int {
	return int(float64(t.count) / float64(len(steam.Countries)))
}
