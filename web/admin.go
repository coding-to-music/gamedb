package web

import (
	"encoding/json"
	"net/http"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/99designs/basicauth-go"
	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/queue"
	"github.com/gamedb/website/websockets"
	"github.com/go-chi/chi"
	"github.com/gosimple/slug"
	"github.com/spf13/viper"
)

func adminRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(basicauth.New("Steam", map[string][]string{
		viper.GetString("ADMIN_USER"): {viper.GetString("ADMIN_PASS")},
	}))
	r.Get("/", adminHandler)
	r.Get("/{option}", adminHandler)
	r.Post("/{option}", adminHandler)
	return r
}

func adminHandler(w http.ResponseWriter, r *http.Request) {

	option := chi.URLParam(r, "option")

	switch option {
	case "refresh-all-apps":
		go adminQueueEveryApp()
	case "refresh-all-packages":
		go adminQueueEveryPackage()
	case "refresh-genres":
		go adminGenres()
	case "refresh-tags":
		go adminTags()
	case "refresh-developers":
		go adminDevelopers()
	case "refresh-publishers":
		go adminPublishers()
	case "refresh-donations":
		go adminDonations()
	case "refresh-ranks":
		go adminRanks()
	case "wipe-memcache":
		go adminMemcache()
	case "disable-consumers":
		go adminDisableConsumers()
	case "run-dev-code":
		go adminDev()
	case "queues":
		err := r.ParseForm()
		log.Log(err)
		go adminQueues(r)
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
		db.ConfWipeMemcache,
		db.ConfRunDevCode,
	})
	log.Log(err)

	// Template
	t := adminTemplate{}
	t.Fill(w, r, "Admin", "")
	t.Configs = configs
	t.Goroutines = runtime.NumGoroutine()

	err = returnTemplate(w, r, "admin", t)
	log.Log(err)
}

type adminTemplate struct {
	GlobalTemplate
	Errors     []string
	Configs    map[string]db.Config
	Goroutines int
}

func adminDisableConsumers() {

}

func adminQueueEveryApp() {

	var last = 0
	var keepGoing = true
	var apps steam.AppList
	var err error

	for keepGoing {

		apps, _, err = helpers.GetSteam().GetAppList(1000, last)
		if err != nil {
			log.Log(err)
			return
		}

		for _, v := range apps.Apps {
			err = queue.Produce(queue.QueueApps, []byte(strconv.Itoa(v.AppID)))
			if err != nil {
				log.Log(err)
				return
			}
			last = v.AppID
		}

		keepGoing = apps.HaveMoreResults
	}

	//
	err = db.SetConfig(db.ConfAddedAllApps, strconv.FormatInt(time.Now().Unix(), 10))
	log.Log(err)

	page, err := websockets.GetPage(websockets.PageAdmin)
	log.Log(err)

	page.Send(adminWebsocket{db.ConfAddedAllApps + " complete"})

	log.Info(strconv.Itoa(len(apps.Apps)) + " apps added to rabbit")
}

func adminQueueEveryPackage() {

	apps, err := db.GetAppsWithPackages()
	if err != nil {
		log.Log(err)
		return
	}

	packageIDs := map[int]bool{}
	for _, v := range apps {

		packages, err := v.GetPackages()
		if err != nil {
			log.Log(err)
			return
		}

		for _, vv := range packages {
			packageIDs[vv] = true
		}
	}

	for k := range packageIDs {

		err = queue.Produce(queue.QueuePackages, []byte(strconv.Itoa(k)))
		if err != nil {
			log.Log(err)
			return
		}
	}

	//
	err = db.SetConfig(db.ConfAddedAllPackages, strconv.FormatInt(time.Now().Unix(), 10))
	log.Log(err)

	page, err := websockets.GetPage(websockets.PageAdmin)
	log.Log(err)

	page.Send(adminWebsocket{db.ConfAddedAllPackages + " complete"})

	log.Info(strconv.Itoa(len(packageIDs)) + " packages added to rabbit")
}

func adminDonations() {

	donations, err := db.GetDonations(0, 0)
	if err != nil {
		cronLog(err)
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
			cronLog(err)
			continue
		}

		player.Donated = v
		err = db.SaveKind(player.GetKey(), player)
		cronLog(err)
	}

	//
	err = db.SetConfig(db.ConfDonationsUpdated, strconv.FormatInt(time.Now().Unix(), 10))
	cronLog(err)

	page, err := websockets.GetPage(websockets.PageAdmin)
	log.Log(err)

	page.Send(adminWebsocket{db.ConfDonationsUpdated + " complete"})

	cronLog("Updated " + strconv.Itoa(len(counts)) + " player donation counts")
}

func adminQueues(r *http.Request) {

	if val := r.PostForm.Get("player-id"); val != "" {

		log.Info("Player ID: " + val)

		playerID, err := strconv.ParseInt(val, 10, 64)
		log.Log(err)

		player := db.Player{}
		player.PlayerID = playerID

		err = queue.QueuePlayer(r, player, db.PlayerUpdateAdmin)
		err = helpers.IgnoreErrors(err, db.ErrUpdatingPlayerBot, db.ErrUpdatingPlayerTooSoon, db.ErrUpdatingPlayerInQueue)
		log.Log(err)
	}

	if val := r.PostForm.Get("app-id"); val != "" {

		log.Info("App ID: " + val)
		err := queue.Produce(queue.QueueApps, []byte(val))
		log.Log(err)
	}

	if val := r.PostForm.Get("package-id"); val != "" {

		log.Info("Package ID: " + val)
		err := queue.Produce(queue.QueuePackages, []byte(val))
		log.Log(err)
	}
}

func adminGenres() {

	cronLog(log.ServiceLocal, "Genres updating")

	gorm, err := db.GetMySQLClient()
	if err != nil {
		cronLog(err)
		return
	}

	// Get current genres, to delete old ones
	currentGenres, err := db.GetAllGenres()
	if err != nil {
		cronLog(err)
		return
	}

	genresToDelete := map[int]int{}
	for _, v := range currentGenres {
		genresToDelete[v.ID] = v.ID
	}

	// Get apps from mysql
	appsWithGenres, err := db.GetAppsWithGenres()
	cronLog(err)

	cronLog("Found " + strconv.Itoa(len(appsWithGenres)) + " apps with genres")

	newGenres := make(map[int]*statsRow)
	for _, app := range appsWithGenres {

		appGenres, err := app.GetGenres()
		if err != nil {
			cronLog(err)
			continue
		}

		if len(appGenres) == 0 {
			appGenres = []steam.AppDetailsGenre{{ID: 0, Description: ""}}
		}

		// For each genre in an app
		for _, genre := range appGenres {

			delete(genresToDelete, genre.ID)

			if _, ok := newGenres[genre.ID]; ok {
				newGenres[genre.ID].count++
				newGenres[genre.ID].totalScore += app.ReviewsScore
			} else {
				newGenres[genre.ID] = &statsRow{
					name:       strings.TrimSpace(genre.Description),
					count:      1,
					totalScore: app.ReviewsScore,
					totalPrice: map[steam.CountryCode]int{},
				}
			}

			for code := range steam.Countries {
				price, err := app.GetPrice(code)
				if err != nil {
					//log.Log(err)
					continue
				}
				newGenres[genre.ID].totalPrice[code] += price.Final
			}
		}
	}

	var limit int
	var wg sync.WaitGroup

	// Delete old genres
	limit++
	wg.Add(1)
	go func() {

		defer wg.Done()

		var genresToDeleteSlice []int
		for _, v := range genresToDelete {
			genresToDeleteSlice = append(genresToDeleteSlice, v)
		}

		err := db.DeleteGenres(genresToDeleteSlice)
		cronLog(err)

		limit--

	}()

	// Update current genres
	var count = 1
	for k, v := range newGenres {

		if limit >= 5 {
			wg.Wait()
		}

		adminStatsLogger("genre", count, len(newGenres), v.name)

		limit++
		wg.Add(1)
		go func(genreID int, v *statsRow) {

			defer wg.Done()

			var genre db.Genre

			gorm = gorm.Unscoped().FirstOrInit(&genre, db.Genre{ID: genreID})
			cronLog(gorm.Error)

			genre.Name = v.name
			genre.Apps = v.count
			genre.MeanPrice = v.GetMeanPrice()
			genre.MeanScore = v.GetMeanScore()
			genre.DeletedAt = nil

			gorm = gorm.Unscoped().Save(&genre)
			cronLog(gorm.Error)

			limit--

		}(k, v)

		count++
	}
	wg.Wait()

	//
	err = db.SetConfig(db.ConfGenresUpdated, strconv.FormatInt(time.Now().Unix(), 10))
	cronLog(err)

	page, err := websockets.GetPage(websockets.PageAdmin)
	log.Log(err)

	page.Send(adminWebsocket{db.ConfGenresUpdated + " complete"})

	cronLog("Genres updated")
}

func adminPublishers() {

	cronLog(log.ServiceLocal, "Publishers updating")

	// Get current publishers, to delete old ones
	currentPublishers, err := db.GetAllPublishers()
	if err != nil {
		cronLog(err)
		return
	}

	publishersToDelete := map[string]int{}
	for _, publisherRow := range currentPublishers {
		publishersToDelete[slug.Make(publisherRow.Name)] = publisherRow.ID
	}

	// Get apps from mysql
	appsWithPublishers, err := db.GetAppsWithPublishers()
	cronLog(err)

	cronLog("Found " + strconv.Itoa(len(appsWithPublishers)) + " apps with publishers")

	newPublishers := make(map[string]*statsRow)
	for _, app := range appsWithPublishers {

		appPublishers, err := app.GetPublishers()
		if err != nil {
			cronLog(err)
			continue
		}

		if len(appPublishers) == 0 {
			appPublishers = []string{""}
		}

		// For each publisher in an app
		for _, publisher := range appPublishers {

			delete(publishersToDelete, publisher)

			if _, ok := newPublishers[publisher]; ok {
				newPublishers[publisher].count++
				newPublishers[publisher].totalScore += app.ReviewsScore
			} else {
				newPublishers[publisher] = &statsRow{
					name:       strings.TrimSpace(publisher),
					count:      1,
					totalPrice: map[steam.CountryCode]int{},
					totalScore: app.ReviewsScore,
				}
			}

			for code := range steam.Countries {
				price, err := app.GetPrice(code)
				if err != nil {
					//log.Log(err)
					continue
				}
				newPublishers[publisher].totalPrice[code] += price.Final
			}
		}
	}

	var limit int
	var wg sync.WaitGroup

	// Delete old publishers
	limit++
	wg.Add(1)
	go func() {

		defer wg.Done()

		var pubsToDeleteSlice []int
		for _, v := range publishersToDelete {
			pubsToDeleteSlice = append(pubsToDeleteSlice, v)
		}

		err := db.DeletePublishers(pubsToDeleteSlice)
		cronLog(err)

		limit--

	}()

	gorm, err := db.GetMySQLClient(true)
	if err != nil {
		cronLog(err)
		return
	}

	// Update current publishers
	var count = 1
	for k, v := range newPublishers {

		if limit >= 5 {
			wg.Wait()
		}

		adminStatsLogger("publisher", count, len(newPublishers), k)

		limit++
		wg.Add(1)
		go func(publisherName string, v *statsRow) {

			defer wg.Done()

			var publisher db.Publisher

			gorm = gorm.Unscoped().FirstOrInit(&publisher, db.Publisher{Name: strings.TrimSpace(publisherName)})
			cronLog(gorm.Error)

			publisher.Name = v.name
			publisher.Apps = v.count
			publisher.MeanPrice = v.GetMeanPrice()
			publisher.MeanScore = v.GetMeanScore()
			publisher.DeletedAt = nil

			gorm = gorm.Unscoped().Save(&publisher)
			cronLog(gorm.Error)

			limit--

		}(k, v)

		count++
	}

	wg.Wait()

	err = db.SetConfig(db.ConfPublishersUpdated, strconv.FormatInt(time.Now().Unix(), 10))
	cronLog(err)

	page, err := websockets.GetPage(websockets.PageAdmin)
	log.Log(err)

	page.Send(adminWebsocket{db.ConfPublishersUpdated + " complete"})

	cronLog("Publishers updated")
}

func adminDevelopers() {

	cronLog(log.ServiceLocal, "Developers updating")

	gorm, err := db.GetMySQLClient()
	if err != nil {
		cronLog(err)
		return
	}

	// Get current developers, to delete old ones
	currentDevelopers, err := db.GetAllPublishers()
	if err != nil {
		cronLog(err)
		return
	}

	developersToDelete := map[string]int{}
	for _, v := range currentDevelopers {
		developersToDelete[v.Name] = v.ID
	}

	// Get apps from mysql
	appsWithDevelopers, err := db.GetAppsWithDevelopers()
	cronLog(err)

	cronLog("Found " + strconv.Itoa(len(appsWithDevelopers)) + " apps with developers")

	newDevelopers := make(map[string]*statsRow)
	for _, app := range appsWithDevelopers {

		appDevelopers, err := app.GetDevelopers()
		if err != nil {
			cronLog(err)
			continue
		}

		if len(appDevelopers) == 0 {
			appDevelopers = []string{""}
		}

		// For each developer in an app
		for _, developer := range appDevelopers {

			delete(developersToDelete, developer)

			if _, ok := newDevelopers[developer]; ok {
				newDevelopers[developer].count++
				newDevelopers[developer].totalScore += app.ReviewsScore
			} else {
				newDevelopers[developer] = &statsRow{
					name:       strings.TrimSpace(developer),
					count:      1,
					totalPrice: map[steam.CountryCode]int{},
					totalScore: app.ReviewsScore,
				}
			}

			for code := range steam.Countries {
				price, err := app.GetPrice(code)
				if err != nil {
					//log.Log(err)
					continue
				}
				newDevelopers[developer].totalPrice[code] += price.Final
			}
		}
	}

	var limit int
	var wg sync.WaitGroup

	// Delete old developers
	limit++
	wg.Add(1)
	go func() {

		defer wg.Done()

		var devsToDeleteSlice []int
		for _, v := range developersToDelete {
			devsToDeleteSlice = append(devsToDeleteSlice, v)
		}

		err := db.DeleteDevelopers(devsToDeleteSlice)
		cronLog(err)

		limit--

	}()

	// Update current developers
	var count = 1
	for k, v := range newDevelopers {

		if limit >= 5 {
			wg.Wait()
		}

		adminStatsLogger("developer", count, len(newDevelopers), k)

		limit++
		wg.Add(1)
		go func(developerName string, v *statsRow) {

			defer wg.Done()

			var developer db.Developer

			gorm = gorm.Unscoped().FirstOrInit(&developer, db.Developer{Name: strings.TrimSpace(developerName)})
			cronLog(gorm.Error)

			developer.Name = v.name
			developer.Apps = v.count
			developer.MeanPrice = v.GetMeanPrice()
			developer.MeanScore = v.GetMeanScore()
			developer.DeletedAt = nil

			gorm = gorm.Unscoped().Save(&developer)
			cronLog(gorm.Error)

			limit--

		}(k, v)

		count++
	}
	wg.Wait()

	err = db.SetConfig(db.ConfDevelopersUpdated, strconv.FormatInt(time.Now().Unix(), 10))
	cronLog(err)

	page, err := websockets.GetPage(websockets.PageAdmin)
	log.Log(err)

	page.Send(adminWebsocket{db.ConfDevelopersUpdated + " complete"})

	cronLog("Developers updated")
}

func adminTags() {

	gorm, err := db.GetMySQLClient()
	if err != nil {
		cronLog(err)
		return
	}

	// Get current tags, to delete old ones
	tags, err := db.GetAllTags()
	if err != nil {
		cronLog(err)
		return
	}

	tagsToDelete := map[int]int{}
	for _, tag := range tags {
		tagsToDelete[tag.ID] = tag.ID
	}

	// Get tags from Steam
	tagsResp, _, err := helpers.GetSteam().GetTags()
	cronLog(err)

	steamTagMap := tagsResp.GetMap()

	appsWithTags, err := db.GetAppsWithTags()
	cronLog(err)

	cronLog("Found " + strconv.Itoa(len(appsWithTags)) + " apps with tags")

	newTags := make(map[int]*statsRow)
	for _, app := range appsWithTags {

		appTags, err := app.GetTagIDs()
		if err != nil {
			cronLog(err)
			continue
		}

		if len(appTags) == 0 {
			//appTags = []int{}
		}

		// For each tag in an app
		for _, tagID := range appTags {

			delete(tagsToDelete, tagID)

			if _, ok := newTags[tagID]; ok {
				newTags[tagID].count++
				newTags[tagID].totalScore += app.ReviewsScore
			} else {
				newTags[tagID] = &statsRow{
					name:       strings.TrimSpace(steamTagMap[tagID]),
					count:      1,
					totalPrice: map[steam.CountryCode]int{},
					totalScore: app.ReviewsScore,
				}
			}

			for code := range steam.Countries {
				price, err := app.GetPrice(code)
				if err != nil {
					//log.Log(err)
					continue
				}
				newTags[tagID].totalPrice[code] += price.Final
			}
		}
	}

	var limit int
	var wg sync.WaitGroup

	// Delete old tags
	limit++
	wg.Add(1)
	go func() {

		defer wg.Done()

		var tagsToDeleteSlice []int
		for _, v := range tagsToDelete {
			tagsToDeleteSlice = append(tagsToDeleteSlice, v)
		}

		err := db.DeleteTags(tagsToDeleteSlice)
		cronLog(err)

		limit--

	}()

	// Update current tags
	var count = 1
	for k, v := range newTags {

		if limit >= 5 {
			wg.Wait()
		}

		adminStatsLogger("tag", count, len(newTags), strconv.Itoa(k))

		limit++
		wg.Add(1)
		go func(tagID int, v *statsRow) {

			defer wg.Done()

			var tag db.Tag

			gorm = gorm.Unscoped().FirstOrInit(&tag, db.Tag{ID: tagID})
			cronLog(gorm.Error)

			tag.Name = v.name
			tag.Apps = v.count
			tag.MeanPrice = v.GetMeanPrice()
			tag.MeanScore = v.GetMeanScore()
			tag.DeletedAt = nil

			gorm = gorm.Unscoped().Save(&tag)
			cronLog(gorm.Error)

			limit--

		}(k, v)

		count++
	}
	wg.Wait()

	err = db.SetConfig(db.ConfTagsUpdated, strconv.FormatInt(time.Now().Unix(), 10))
	cronLog(err)

	page, err := websockets.GetPage(websockets.PageAdmin)
	log.Log(err)

	page.Send(adminWebsocket{db.ConfTagsUpdated + " complete"})

	cronLog("Tags updated")
}

func adminStatsLogger(tableName string, count int, total int, rowName string) {

	log.Info("Updating " + tableName + " - " + strconv.Itoa(count) + " / " + strconv.Itoa(total) + ": " + rowName)
}

func adminRanks() {

	cronLog("Ranks updated started")

	timeStart := time.Now().Unix()

	oldKeys, err := db.GetRankKeys()
	if err != nil {
		cronLog(err)
		return
	}

	newRanks := make(map[int64]*db.PlayerRank)
	var players []db.Player

	var wg sync.WaitGroup

	for _, v := range []string{"-level", "-games_count", "-badges_count", "-play_time", "-friends_count"} {

		wg.Add(1)
		go func(column string) {

			defer wg.Done()

			players, err = db.GetAllPlayers(column, db.PlayersToRank)
			if err != nil {
				cronLog(err)
				return
			}

			for _, v := range players {
				newRanks[v.PlayerID] = db.NewRankFromPlayer(v)
				delete(oldKeys, v.PlayerID)
			}

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
		cronLog(err)
		return
	}

	// Remove old ranks
	var keysToDelete []*datastore.Key
	for _, v := range oldKeys {
		keysToDelete = append(keysToDelete, v)
	}

	err = db.BulkDeleteKinds(keysToDelete, false)
	if err != nil {
		cronLog(err)
		return
	}

	// Update config
	err = db.SetConfig(db.ConfRanksUpdated, strconv.FormatInt(time.Now().Unix(), 10))
	cronLog(err)

	page, err := websockets.GetPage(websockets.PageAdmin)
	page.Send(adminWebsocket{db.ConfRanksUpdated + " complete"})

	cronLog("Ranks updated in " + strconv.FormatInt(time.Now().Unix()-timeStart, 10) + " seconds")
}

func adminMemcache() {

	err := helpers.GetMemcache().DeleteAll()
	log.Log(err)

	err = db.SetConfig(db.ConfWipeMemcache, strconv.FormatInt(time.Now().Unix(), 10))
	log.Log(err)

	page, err := websockets.GetPage(websockets.PageAdmin)
	log.Log(err)

	page.Send(adminWebsocket{db.ConfWipeMemcache + " complete"})

	log.Info("Memcache wiped")
}

func adminDev() {

	//gorm, err := db.GetMySQLClient()
	//if err != nil {
	//
	//	log.Log(err)
	//	return
	//}
	//
	//var packages []db.App
	//
	//gorm = gorm.Select([]string{"id"})
	//gorm = gorm.Where("name = ?", "appinfo")
	//gorm = gorm.Find(&packages)
	//
	//for _, v := range packages {
	//	err := queue.Produce(queue.QueueApps, []byte(strconv.Itoa(v.ID)))
	//	log.Log(err)
	//}

	// ######################################################

	//log.Info("Running...")
	//
	//client, ctx, err := db.GetDSClient()
	//log.Log(err)
	//
	//q := datastore.NewQuery(db.KindNews)
	//
	//var articles []db.News
	//_, err = client.GetAll(ctx, q, &articles)
	//log.Log(err)
	//
	//var articlesToDelete []*datastore.Key
	//for _, v := range articles {
	//	if strings.TrimSpace(v.Contents) == "" {
	//		articlesToDelete = append(articlesToDelete, v.GetKey())
	//		fmt.Println(v.ArticleID)
	//	}
	//}
	//
	//err = db.BulkDeleteKinds(articlesToDelete, true)
	//log.Log(err)

	// ######################################################

	//log.Info("Dev")
	//
	//players, err := db.GetAllPlayers("__key__", 0)
	//
	//log.Info("Got players")
	//
	//if err != nil {
	//
	//	log.Log(err)
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
	//	log.Log(err)
	//}
	//
	//log.Info("Done")

	err := db.SetConfig(db.ConfRunDevCode, strconv.FormatInt(time.Now().Unix(), 10))
	log.Log(err)

	page, err := websockets.GetPage(websockets.PageAdmin)
	log.Log(err)

	page.Send(adminWebsocket{db.ConfRunDevCode + " complete"})

	log.Info("Dev code run")
}

type statsRow struct {
	name       string
	count      int
	totalPrice map[steam.CountryCode]int
	totalScore float64
}

func (t statsRow) GetMeanPrice() string {

	means := map[steam.CountryCode]float64{}

	for code, total := range t.totalPrice {
		means[code] = float64(total) / float64(t.count)
	}

	bytes, err := json.Marshal(means)
	log.Log(err)

	return string(bytes)
}

func (t statsRow) GetMeanScore() float64 {
	return t.totalScore / float64(t.count)
}

type adminWebsocket struct {
	Message string `json:"message"`
}

func cronLog(interfaces ...interface{}) {
	log.Info(append(interfaces, log.LogNameCron)...)
}
