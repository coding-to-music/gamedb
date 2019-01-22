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
	"github.com/gamedb/website/config"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/queue"
	"github.com/gamedb/website/websockets"
	"github.com/go-chi/chi"
	"github.com/gosimple/slug"
)

func adminRouter() http.Handler {
	r := chi.NewRouter()
	if !config.Config.IsLocal() {
		r.Use(basicauth.New("Steam", map[string][]string{
			config.Config.AdminUsername: {config.Config.AdminPassword},
		}))
	}
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
		log.Err(err, r)
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
		db.ConfWipeMemcache + "-" + config.Config.Environment.Get(),
		db.ConfRunDevCode,
	})
	log.Err(err, r)

	// Template
	t := adminTemplate{}
	t.Fill(w, r, "Admin", "")
	t.Configs = configs
	t.Goroutines = runtime.NumGoroutine()

	err = returnTemplate(w, r, "admin", t)
	log.Err(err, r)
}

type adminTemplate struct {
	GlobalTemplate
	Errors     []string
	Configs    map[string]db.Config
	Goroutines int
}

func (at adminTemplate) GetMCConfigKey() string {
	return "wipe-memcache" + "-" + config.Config.Environment.Get()
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
			log.Err(err)
			return
		}

		for _, v := range apps.Apps {
			err = queue.QueueApp([]int{v.AppID})
			if err != nil {
				log.Err(err)
				return
			}
			last = v.AppID
		}

		keepGoing = apps.HaveMoreResults
	}

	//
	err = db.SetConfig(db.ConfAddedAllApps, strconv.FormatInt(time.Now().Unix(), 10))
	log.Err(err)

	page, err := websockets.GetPage(websockets.PageAdmin)
	log.Err(err)

	page.Send(adminWebsocket{db.ConfAddedAllApps + " complete"})

	log.Info(strconv.Itoa(len(apps.Apps)) + " apps added to rabbit")
}

func adminQueueEveryPackage() {

	apps, err := db.GetAppsWithPackages()
	if err != nil {
		log.Err(err)
		return
	}

	packageIDs := map[int]bool{}
	for _, v := range apps {

		packages, err := v.GetPackages()
		if err != nil {
			log.Err(err)
			return
		}

		for _, vv := range packages {
			packageIDs[vv] = true
		}
	}

	for k := range packageIDs {

		err = queue.QueuePackage([]int{k})
		if err != nil {
			log.Err(err)
			return
		}
	}

	//
	err = db.SetConfig(db.ConfAddedAllPackages, strconv.FormatInt(time.Now().Unix(), 10))
	log.Err(err)

	page, err := websockets.GetPage(websockets.PageAdmin)
	log.Err(err)

	page.Send(adminWebsocket{db.ConfAddedAllPackages + " complete"})

	log.Info(strconv.Itoa(len(packageIDs)) + " packages added to rabbit")
}

func adminDonations() {

	donations, err := db.GetDonations(0, 0)
	if err != nil {
		cronLogErr(err)
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
			cronLogErr(err)
			continue
		}

		player.Donated = v
		err = db.SaveKind(player.GetKey(), player)
		cronLogErr(err)
	}

	//
	err = db.SetConfig(db.ConfDonationsUpdated, strconv.FormatInt(time.Now().Unix(), 10))
	cronLogErr(err)

	page, err := websockets.GetPage(websockets.PageAdmin)
	log.Err(err)

	page.Send(adminWebsocket{db.ConfDonationsUpdated + " complete"})

	cronLogInfo("Updated " + strconv.Itoa(len(counts)) + " player donation counts")
}

func adminQueues(r *http.Request) {

	if val := r.PostForm.Get("player-id"); val != "" {

		playerID, err := strconv.ParseInt(val, 10, 64)
		log.Err(err, r)

		err = queue.QueuePlayer(playerID)
		log.Err(err, r)
	}

	if val := r.PostForm.Get("app-id"); val != "" {

		valInt, err := strconv.ParseInt(val, 10, 32)
		log.Err(err, r)

		err = queue.QueueApp([]int{int(valInt)})
		log.Err(err, r)
	}

	if val := r.PostForm.Get("package-id"); val != "" {

		valInt, err := strconv.ParseInt(val, 10, 32)
		log.Err(err, r)

		err = queue.QueuePackage([]int{int(valInt)})
		log.Err(err, r)
	}

	if val := r.PostForm.Get("bundle-id"); val != "" {

		valInt, err := strconv.ParseInt(val, 10, 32)
		log.Err(err, r)

		err = queue.QueueBundle(int(valInt))
		log.Err(err, r)
	}
}

func adminGenres() {

	cronLogInfo(log.ServiceLocal, "Genres updating")

	gorm, err := db.GetMySQLClient()
	if err != nil {
		cronLogErr(err)
		return
	}

	// Get current genres, to delete old ones
	currentGenres, err := db.GetAllGenres()
	if err != nil {
		cronLogErr(err)
		return
	}

	genresToDelete := map[int]int{}
	for _, v := range currentGenres {
		genresToDelete[v.ID] = v.ID
	}

	// Get apps from mysql
	appsWithGenres, err := db.GetAppsWithGenres()
	cronLogErr(err)

	cronLogInfo("Found " + strconv.Itoa(len(appsWithGenres)) + " apps with genres")

	newGenres := make(map[int]*statsRow)
	for _, app := range appsWithGenres {

		appGenres, err := app.GetGenres()
		if err != nil {
			cronLogErr(err)
			continue
		}

		if len(appGenres) == 0 {
			appGenres = []steam.AppDetailsGenre{{ID: "0", Description: ""}}
		}

		// For each genre in an app
		for _, genre := range appGenres {

			genreID, err := strconv.Atoi(genre.ID)
			if err != nil {
				cronLogErr(err)
				continue
			}

			delete(genresToDelete, genreID)

			if _, ok := newGenres[genreID]; ok {
				newGenres[genreID].count++
				newGenres[genreID].totalScore += app.ReviewsScore
			} else {
				newGenres[genreID] = &statsRow{
					name:       strings.TrimSpace(genre.Description),
					count:      1,
					totalScore: app.ReviewsScore,
					totalPrice: map[steam.CountryCode]int{},
				}
			}

			for code := range steam.Countries {
				price, err := app.GetPrice(code)
				if err != nil {
					// cronLogErr(err, r)
					continue
				}
				newGenres[genreID].totalPrice[code] += price.Final
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
		cronLogErr(err)

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
			cronLogErr(gorm.Error)

			genre.Name = v.name
			genre.Apps = v.count
			genre.MeanPrice = v.GetMeanPrice()
			genre.MeanScore = v.GetMeanScore()
			genre.DeletedAt = nil

			gorm = gorm.Unscoped().Save(&genre)
			cronLogErr(gorm.Error)

			limit--

		}(k, v)

		count++
	}
	wg.Wait()

	//
	err = db.SetConfig(db.ConfGenresUpdated, strconv.FormatInt(time.Now().Unix(), 10))
	cronLogErr(err)

	page, err := websockets.GetPage(websockets.PageAdmin)
	cronLogErr(err)

	page.Send(adminWebsocket{db.ConfGenresUpdated + " complete"})

	cronLogInfo("Genres updated")
}

func adminPublishers() {

	cronLogInfo(log.ServiceLocal, "Publishers updating")

	// Get current publishers, to delete old ones
	currentPublishers, err := db.GetAllPublishers()
	if err != nil {
		cronLogErr(err)
		return
	}

	publishersToDelete := map[string]int{}
	for _, publisherRow := range currentPublishers {
		publishersToDelete[slug.Make(publisherRow.Name)] = publisherRow.ID
	}

	// Get apps from mysql
	appsWithPublishers, err := db.GetAppsWithPublishers()
	cronLogErr(err)

	cronLogInfo("Found " + strconv.Itoa(len(appsWithPublishers)) + " apps with publishers")

	newPublishers := make(map[string]*statsRow)
	for _, app := range appsWithPublishers {

		appPublishers, err := app.GetPublishers()
		if err != nil {
			cronLogErr(err)
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
					// cronLogErr(err, r)
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
		cronLogErr(err)

		limit--

	}()

	gorm, err := db.GetMySQLClient(true)
	if err != nil {
		cronLogErr(err)
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
			cronLogErr(gorm.Error)

			publisher.Name = v.name
			publisher.Apps = v.count
			publisher.MeanPrice = v.GetMeanPrice()
			publisher.MeanScore = v.GetMeanScore()
			publisher.DeletedAt = nil

			gorm = gorm.Unscoped().Save(&publisher)
			cronLogErr(gorm.Error)

			limit--

		}(k, v)

		count++
	}

	wg.Wait()

	err = db.SetConfig(db.ConfPublishersUpdated, strconv.FormatInt(time.Now().Unix(), 10))
	cronLogErr(err)

	page, err := websockets.GetPage(websockets.PageAdmin)
	cronLogErr(err)

	page.Send(adminWebsocket{db.ConfPublishersUpdated + " complete"})

	cronLogInfo("Publishers updated")
}

func adminDevelopers() {

	cronLogInfo(log.ServiceLocal, "Developers updating")

	gorm, err := db.GetMySQLClient()
	if err != nil {
		cronLogErr(err)
		return
	}

	// Get current developers, to delete old ones
	currentDevelopers, err := db.GetAllPublishers()
	if err != nil {
		cronLogErr(err)
		return
	}

	developersToDelete := map[string]int{}
	for _, v := range currentDevelopers {
		developersToDelete[v.Name] = v.ID
	}

	// Get apps from mysql
	appsWithDevelopers, err := db.GetAppsWithDevelopers()
	cronLogErr(err)

	cronLogErr("Found " + strconv.Itoa(len(appsWithDevelopers)) + " apps with developers")

	newDevelopers := make(map[string]*statsRow)
	for _, app := range appsWithDevelopers {

		appDevelopers, err := app.GetDevelopers()
		if err != nil {
			cronLogErr(err)
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
					// cronLogErr(err, r)
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
		cronLogErr(err)

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
			cronLogErr(gorm.Error)

			developer.Name = v.name
			developer.Apps = v.count
			developer.MeanPrice = v.GetMeanPrice()
			developer.MeanScore = v.GetMeanScore()
			developer.DeletedAt = nil

			gorm = gorm.Unscoped().Save(&developer)
			cronLogErr(gorm.Error)

			limit--

		}(k, v)

		count++
	}
	wg.Wait()

	err = db.SetConfig(db.ConfDevelopersUpdated, strconv.FormatInt(time.Now().Unix(), 10))
	cronLogErr(err)

	page, err := websockets.GetPage(websockets.PageAdmin)
	cronLogErr(err)

	page.Send(adminWebsocket{db.ConfDevelopersUpdated + " complete"})

	cronLogInfo("Developers updated")
}

func adminTags() {

	gorm, err := db.GetMySQLClient()
	if err != nil {
		cronLogErr(err)
		return
	}

	// Get current tags, to delete old ones
	tags, err := db.GetAllTags()
	if err != nil {
		cronLogErr(err)
		return
	}

	tagsToDelete := map[int]int{}
	for _, tag := range tags {
		tagsToDelete[tag.ID] = tag.ID
	}

	// Get tags from Steam
	tagsResp, _, err := helpers.GetSteam().GetTags()
	cronLogErr(err)

	steamTagMap := tagsResp.GetMap()

	appsWithTags, err := db.GetAppsWithTags()
	cronLogErr(err)

	cronLogErr("Found " + strconv.Itoa(len(appsWithTags)) + " apps with tags")

	newTags := make(map[int]*statsRow)
	for _, app := range appsWithTags {

		appTags, err := app.GetTagIDs()
		if err != nil {
			cronLogErr(err)
			continue
		}

		if len(appTags) == 0 {
			// appTags = []int{}
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
					// cronLogErr(err, r)
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
		cronLogErr(err)

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
			cronLogErr(gorm.Error)

			tag.Name = v.name
			tag.Apps = v.count
			tag.MeanPrice = v.GetMeanPrice()
			tag.MeanScore = v.GetMeanScore()
			tag.DeletedAt = nil

			gorm = gorm.Unscoped().Save(&tag)
			cronLogErr(gorm.Error)

			limit--

		}(k, v)

		count++
	}
	wg.Wait()

	err = db.SetConfig(db.ConfTagsUpdated, strconv.FormatInt(time.Now().Unix(), 10))
	cronLogErr(err)

	page, err := websockets.GetPage(websockets.PageAdmin)
	cronLogErr(err)

	page.Send(adminWebsocket{db.ConfTagsUpdated + " complete"})

	cronLogInfo("Tags updated")
}

func adminStatsLogger(tableName string, count int, total int, rowName string) {

	log.Info("Updating " + tableName + " - " + strconv.Itoa(count) + " / " + strconv.Itoa(total) + ": " + rowName)
}

func adminRanks() {

	cronLogErr("Ranks updated started")

	timeStart := time.Now().Unix()

	oldKeys, err := db.GetRankKeys()
	if err != nil {
		cronLogErr(err)
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
				cronLogErr(err)
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
		cronLogErr(err)
		return
	}

	// Remove old ranks
	var keysToDelete []*datastore.Key
	for _, v := range oldKeys {
		keysToDelete = append(keysToDelete, v)
	}

	err = db.BulkDeleteKinds(keysToDelete, false)
	if err != nil {
		cronLogErr(err)
		return
	}

	// Update config
	err = db.SetConfig(db.ConfRanksUpdated, strconv.FormatInt(time.Now().Unix(), 10))
	cronLogErr(err)

	page, err := websockets.GetPage(websockets.PageAdmin)
	page.Send(adminWebsocket{db.ConfRanksUpdated + " complete"})

	cronLogInfo("Ranks updated in " + strconv.FormatInt(time.Now().Unix()-timeStart, 10) + " seconds")
}

func adminMemcache() {

	err := helpers.GetMemcache().DeleteAll()
	log.Err(err)

	err = db.SetConfig(db.ConfWipeMemcache+"-"+config.Config.Environment.Get(), strconv.FormatInt(time.Now().Unix(), 10))
	log.Err(err)

	page, err := websockets.GetPage(websockets.PageAdmin)
	log.Err(err)

	page.Send(adminWebsocket{db.ConfWipeMemcache + "-" + config.Config.Environment.Get() + " complete"})

	log.Info("Memcache wiped")
}

func adminDev() {

	var err error

	// gorm, err := db.GetMySQLClient()
	// if err != nil {
	// 	log.Err(err)
	// 	return
	// }
	//
	// var apps []db.App
	//
	// gorm = gorm.Select([]string{"id"})
	// gorm = gorm.Limit(10000)
	// gorm = gorm.Find(&apps)
	//
	// fmt.Println("Found " + humanize.Comma(int64(len(apps))) + "apps")
	//
	// var wg = sync.WaitGroup{}
	// var count int
	// for _, v := range apps {
	//
	// 	wg.Add(1)
	// 	go func(v db.App) {
	//
	// 		defer wg.Done()
	//
	// 		players, _, err := helpers.GetSteam().GetNumberOfCurrentPlayers(v.ID)
	//
	// 		err2, ok := err.(steam.Error)
	// 		if ok && (err2.Code() == 404) {
	// 			fmt.Println("-")
	// 			return
	// 		}
	//
	// 		if err != nil {
	// 			fmt.Println(err)
	// 			return
	// 		}
	//
	// 		if players > 0 {
	// 			fmt.Println(players)
	// 			count++
	// 		}
	//
	// 	}(v)
	// }
	//
	// wg.Wait()
	//
	// fmt.Println(strconv.Itoa(count) + " apps with players")

	// ######################################################

	gorm, err := db.GetMySQLClient()
	if err != nil {

		log.Err(err)
		return
	}

	var packages []db.App

	gorm = gorm.Select([]string{"id"})
	gorm = gorm.Where("achievements LIKE ?", "{%")
	gorm = gorm.Limit(1000)
	gorm = gorm.Order("reviews_score desc")
	gorm = gorm.Find(&packages)

	for _, v := range packages {
		err = queue.QueueApp([]int{v.ID})
		log.Err(err)
	}

	// ######################################################

	// log.Info("Running...")
	//
	// client, ctx, err := db.GetDSClient()
	// log.Err(err, r)
	//
	// q := datastore.NewQuery(db.KindNews)
	//
	// var articles []db.News
	// _, err = client.GetAll(ctx, q, &articles)
	// log.Err(err, r)
	//
	// var articlesToDelete []*datastore.Key
	// for _, v := range articles {
	// 	if strings.TrimSpace(v.Contents) == "" {
	// 		articlesToDelete = append(articlesToDelete, v.GetKey())
	// 		fmt.Println(v.ArticleID)
	// 	}
	// }
	//
	// err = db.BulkDeleteKinds(articlesToDelete, true)
	// log.Err(err, r)

	// ######################################################

	// log.Info("Dev")
	//
	// players, err := db.GetAllPlayers("__key__", 0)
	//
	// log.Info("Got players")
	//
	// if err != nil {
	//
	// 	log.Err(err, r)
	//
	// 	if _, ok := err.(*ds.ErrFieldMismatch); ok {
	//
	// 	} else {
	// 		return
	// 	}
	// }
	//
	// for _, v := range players {
	// 	//v.Games = ""
	// 	err := v.Save()
	// 	log.Err(err, r)
	// }
	//
	// log.Info("Done")

	err = db.SetConfig(db.ConfRunDevCode, strconv.FormatInt(time.Now().Unix(), 10))
	log.Err(err)

	page, err := websockets.GetPage(websockets.PageAdmin)
	log.Err(err)

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
	log.Err(err)

	return string(bytes)
}

func (t statsRow) GetMeanScore() float64 {
	return t.totalScore / float64(t.count)
}

type adminWebsocket struct {
	Message string `json:"message"`
}

func cronLogErr(interfaces ...interface{}) {
	log.Err(append(interfaces, log.LogNameCron, log.LogNameGameDB)...)
}

func cronLogInfo(interfaces ...interface{}) {
	log.Info(append(interfaces, log.LogNameCron, log.LogNameGameDB)...)
}
