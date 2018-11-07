package web

import (
	"encoding/json"
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

	// Get apps with a genre
	gorm, err := db.GetMySQLClient()
	if err != nil {
		logging.Error(err)
		return
	}

	gorm = gorm.Select([]string{})
	gorm = gorm.Where("JSON_DEPTH(genres) = ?", 3)

	var apps []db.App

	gorm = gorm.Find(&apps)
	if gorm.Error != nil {
		logging.Error(gorm.Error)
		return
	}

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

	logging.InfoL("Publishers Updating")

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

	// Get apps with a publisher
	gorm, err := db.GetMySQLClient()
	if err != nil {
		logging.Error(err)
		return
	}

	gorm = gorm.Select([]string{"name", "price_final", "publishers", "reviews_score"})
	// todo, filter on apps with a publisher

	var apps []db.App

	gorm = gorm.Find(&apps)
	if gorm.Error != nil {
		logging.Error(gorm.Error)
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

			price := app.GetPrice(steam.CountryUS) // todo, need to do this for all codes?

			if _, ok := publishersToAdd[key]; ok {
				publishersToAdd[key].count++
				publishersToAdd[key].totalPrice += price.Final
				publishersToAdd[key].totalScore += app.ReviewsScore
			} else {
				publishersToAdd[key] = &adminPublisher{
					count:      1,
					totalPrice: price.Final,
					totalScore: app.ReviewsScore,
					name:       publisher,
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
		for _, v := range pubsToDelete {
			pubsToDeleteSlice = append(pubsToDeleteSlice, v)
		}

		err := db.DeletePublishers(pubsToDeleteSlice)
		logging.Error(err)

		limit--
		wg.Done()
	}()

	// Update current publishers
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
		key := slug.Make(v.Name)
		devsToDelete[key] = v.ID
	}

	// Get apps from mysql
	apps, err := db.GetAppsWithDevelopers()
	logging.Error(err)

	counts := make(map[string]*adminDeveloper)

	for _, app := range apps {

		appDevelopers, err := app.GetDevelopers()
		if err != nil {
			logging.Error(err)
			continue
		}

		if len(appDevelopers) == 0 {
			appDevelopers = []string{"No Developer"}
		}

		for _, developer := range appDevelopers {

			key := slug.Make(developer)

			delete(devsToDelete, key)

			price := app.GetPrice(steam.CountryUS) // todo, need to do this for all codes?

			if _, ok := counts[key]; ok {
				counts[key].count++
				counts[key].totalPrice += price.Final
				counts[key].totalScore += app.ReviewsScore
			} else {
				counts[key] = &adminDeveloper{
					name:       app.GetName(),
					count:      1,
					totalPrice: price.Final,
					totalScore: app.ReviewsScore,
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
		for _, v := range devsToDelete {
			devsToDeleteSlice = append(devsToDeleteSlice, v)
		}

		db.DeleteDevelopers(devsToDeleteSlice)

		limit--
		wg.Done()
	}()

	// Update current developers
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

	newTags := make(map[int]*adminTag)
	for _, app := range appsWithTags {

		tags, err := app.GetTagIDs()
		if err != nil {
			logging.Error(err)
			continue
		}

		for _, tagID := range tags {

			delete(tagsToDelete, tagID)

			for code := range steam.Countries {

				price := app.GetPrice(code)

				if _, ok := newTags[tagID]; ok {
					newTags[tagID].count++
					newTags[tagID].totalPrice[code] += price.Final
					newTags[tagID].totalScore[code] += app.ReviewsScore
				} else {
					newTags[tagID] = &adminTag{
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
		go func(tagID int, v *adminTag) {

			tag := db.Tag{
				ID:        tagID,
				Name:      v.name,
				Apps:      v.count,
				MeanPrice: v.GetMeanPrice(),
				MeanScore: v.GetMeanScore(),
			}

			gorm = gorm.Where(db.Tag{ID: tagID}).Assign(v).FirstOrCreate(&tag)
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

type adminTag struct {
	name       string
	count      int
	totalPrice map[steam.CountryCode]int
	totalScore map[steam.CountryCode]float64
}

func (t adminTag) GetMeanPrice() string {

	means := map[steam.CountryCode]float64{}

	for code, total := range t.totalPrice {
		means[code] = float64(total) / float64(t.count)
	}

	bytes, err := json.Marshal(means)
	logging.Error(err)

	return string(bytes)
}

func (t adminTag) GetMeanScore() string {

	means := map[steam.CountryCode]float64{}

	for code, total := range t.totalScore {
		means[code] = float64(total) / float64(t.count)
	}

	bytes, err := json.Marshal(means)
	logging.Error(err)

	return string(bytes)
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
