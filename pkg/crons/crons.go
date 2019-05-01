package crons

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Jleagle/steam-go/steam"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/websockets"
	influx "github.com/influxdata/influxdb1-client"
)

func AutoPlayerRefreshes() {

	cronLogInfo("Running auto profile updates")

	gorm, err := sql.GetMySQLClient()
	if err != nil {
		log.Err(err)
		return
	}

	gorm = gorm.Select([]string{"player_id"})
	gorm = gorm.Where("patreon_level >= ?", 3)

	var users []sql.User
	gorm = gorm.Find(&users)
	if gorm.Error != nil {
		log.Err(gorm.Error)
		return
	}

	for _, v := range users {
		err := queue.ProducePlayer(v.PlayerID)
		log.Err(err)
	}
}

func Instagram() {

	log.Info("Running IG")

	gorm, err := sql.GetMySQLClient()
	if err != nil {
		log.Err(err)
		return
	}

	gorm = gorm.Select([]string{"id", "name", "screenshots", "reviews_score"})
	gorm = gorm.Where("JSON_DEPTH(screenshots) = ?", 3)
	gorm = gorm.Where("name != ?", "")
	gorm = gorm.Where("type = ?", "game")
	gorm = gorm.Where("reviews_score >= ?", 90)
	gorm = gorm.Order("RAND()")
	gorm = gorm.Limit(1)

	var apps []sql.App
	gorm = gorm.Find(&apps)
	if gorm.Error != nil {
		log.Err(gorm.Error)
		return
	}

	if len(apps) == 0 {
		log.Err("no apps found for instagram")
		return
	}

	var app = apps[0]

	screenshots, err := app.GetScreenshots()
	if err != nil {
		log.Err(err)
		return
	}

	var url = screenshots[rand.Intn(len(screenshots))].PathFull
	if url == "" {
		Instagram()
		return
	}

	err = helpers.UploadInstagram(url, app.GetName()+" (Score: "+helpers.FloatToString(app.ReviewsScore, 2)+") https://gamedb.online/apps/"+strconv.Itoa(app.ID)+" #steamgames #steam #gaming "+helpers.GetHashTag(app.GetName()))
	log.Critical(err, url)
}

func Donations() {

	// donations, err := db.GetDonations(0, 0)
	// if err != nil {
	// 	cronLogErr(err)
	// 	return
	// }
	//
	// // map[player]total
	// counts := make(map[int64]int)
	//
	// for _, v := range donations {
	//
	// 	if _, ok := counts[v.PlayerID]; ok {
	// 		counts[v.PlayerID] = counts[v.PlayerID] + v.AmountUSD
	// 	} else {
	// 		counts[v.PlayerID] = v.AmountUSD
	// 	}
	// }
	//
	// for k, v := range counts {
	// 	player, err := mongo.GetPlayer(k)
	// 	if err != nil {
	// 		cronLogErr(err)
	// 		continue
	// 	}
	//
	// 	player.Donated = v
	// 	err = db.SaveKind(player.GetKey(), player)
	// 	cronLogErr(err)
	// }

	//
	err := sql.SetConfig(sql.ConfDonationsUpdated, strconv.FormatInt(time.Now().Unix(), 10))
	cronLogErr(err)

	page := websockets.GetPage(websockets.PageAdmin)
	page.Send(websockets.AdminPayload{Message: sql.ConfDonationsUpdated + " complete"})

	// cronLogInfo("Updated " + strconv.Itoa(len(counts)) + " player donation counts")
}

func Genres() {

	cronLogInfo("Genres updating")

	// Get current genres, to delete old ones
	currentGenres, err := sql.GetAllGenres(true)
	if err != nil {
		cronLogErr(err)
		return
	}

	genresToDelete := map[int]bool{}
	for _, v := range currentGenres {
		genresToDelete[v.ID] = true
	}

	var genreNameMap = map[int]string{}
	for _, v := range currentGenres {
		genreNameMap[v.ID] = strings.TrimSpace(v.GetName())
	}

	// Get apps from mysql
	appsWithGenres, err := sql.GetAppsWithColumnDepth("genres", 2, []string{"genres", "prices", "reviews_score"})
	cronLogErr(err)

	cronLogInfo("Found " + strconv.Itoa(len(appsWithGenres)) + " apps with genres")

	newGenres := make(map[int]*statsRow)
	for _, app := range appsWithGenres {

		appGenreIDs, err := app.GetGenreIDs()
		if err != nil {
			cronLogErr(err)
			continue
		}

		if len(appGenreIDs) == 0 {
			// appGenreIDs = []db.AppGenre{{ID: 0, Name: ""}}
		}

		// For each genre in an app
		for _, genreID := range appGenreIDs {

			delete(genresToDelete, genreID)

			var genreName string
			if val, ok := genreNameMap[genreID]; ok {
				genreName = val
			} else {
				// genreName = "Unknown"
				continue
			}

			if _, ok := newGenres[genreID]; ok {
				newGenres[genreID].count++
				newGenres[genreID].totalScore += app.ReviewsScore
			} else {
				newGenres[genreID] = &statsRow{
					name:       genreName,
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

		defer func() {
			limit--
			wg.Done()
		}()

		var genresToDeleteSlice []int
		for genreID := range genresToDelete {
			genresToDeleteSlice = append(genresToDeleteSlice, genreID)
		}

		err := sql.DeleteGenres(genresToDeleteSlice)
		cronLogErr(err)

	}()

	wg.Wait()

	gorm, err := sql.GetMySQLClient()
	if err != nil {
		cronLogErr(err)
		return
	}

	// Update current genres
	var count = 1
	for k, v := range newGenres {

		if limit >= 2 {
			wg.Wait()
		}

		statsLogger("genre", count, len(newGenres), v.name)

		limit++
		wg.Add(1)
		go func(genreID int, v *statsRow) {

			defer func() {
				limit--
				wg.Done()
			}()

			var genre sql.Genre

			gorm = gorm.Unscoped().FirstOrInit(&genre, sql.Genre{ID: genreID})
			cronLogErr(gorm.Error)

			genre.Name = v.name
			genre.Apps = v.count
			genre.MeanPrice = v.getMeanPrice()
			genre.MeanScore = v.getMeanScore()
			genre.DeletedAt = nil

			gorm = gorm.Unscoped().Save(&genre)
			cronLogErr(gorm.Error)

		}(k, v)

		count++
	}
	wg.Wait()

	//
	err = sql.SetConfig(sql.ConfGenresUpdated, strconv.FormatInt(time.Now().Unix(), 10))
	cronLogErr(err)

	page := websockets.GetPage(websockets.PageAdmin) //
	page.Send(websockets.AdminPayload{Message: sql.ConfGenresUpdated + " complete"})

	//
	err = helpers.GetMemcache().Delete(helpers.MemcacheGenreKeyNames.Key)
	err = helpers.IgnoreErrors(err, helpers.ErrCacheMiss)
	cronLogErr(err)

	//
	cronLogInfo("Genres updated")
}

func Publishers() {

	cronLogInfo("Publishers updating")

	// Get current publishers, to delete old ones
	allPublishers, err := sql.GetAllPublishers()
	if err != nil {
		cronLogErr(err)
		return
	}

	publishersToDelete := map[int]bool{}
	for _, publisherRow := range allPublishers {
		publishersToDelete[publisherRow.ID] = true
	}

	var publisherNameMap = map[int]string{}
	for _, v := range allPublishers {
		publisherNameMap[v.ID] = strings.TrimSpace(v.GetName())
	}

	// Get apps from mysql
	appsWithPublishers, err := sql.GetAppsWithColumnDepth("publishers", 2, []string{"publishers", "prices", "reviews_score"})
	cronLogErr(err)

	cronLogInfo("Found " + strconv.Itoa(len(appsWithPublishers)) + " apps with publishers")

	newPublishers := make(map[int]*statsRow)
	for _, app := range appsWithPublishers {

		appPublishers, err := app.GetPublisherIDs()
		if err != nil {
			cronLogErr(err)
			continue
		}

		if len(appPublishers) == 0 {
			// appPublishers = []string{""}
		}

		// For each publisher in an app
		for _, appPublisherID := range appPublishers {

			delete(publishersToDelete, appPublisherID)

			var publisherName string
			if val, ok := publisherNameMap[appPublisherID]; ok {
				publisherName = val
			} else {
				// publisherName = "Unknown"
				continue
			}

			if _, ok := newPublishers[appPublisherID]; ok {
				newPublishers[appPublisherID].count++
				newPublishers[appPublisherID].totalScore += app.ReviewsScore
			} else {
				newPublishers[appPublisherID] = &statsRow{
					name:       publisherName,
					count:      1,
					totalPrice: map[steam.CountryCode]int{},
					totalScore: app.ReviewsScore,
				}
			}

			for code := range steam.Countries {
				price, err := app.GetPrice(code)
				if err != nil {
					continue
				}
				newPublishers[appPublisherID].totalPrice[code] += price.Final
			}
		}
	}

	var limit int
	var wg sync.WaitGroup

	// Delete old publishers
	limit++
	wg.Add(1)
	go func() {

		defer func() {
			limit--
			wg.Done()
		}()

		var pubsToDeleteSlice []int
		for publisherID := range publishersToDelete {
			pubsToDeleteSlice = append(pubsToDeleteSlice, publisherID)
		}

		err := sql.DeletePublishers(pubsToDeleteSlice)
		cronLogErr(err)

	}()

	wg.Wait()

	gorm, err := sql.GetMySQLClient()
	if err != nil {
		cronLogErr(err)
		return
	}

	// Update current publishers
	var count = 1
	for k, v := range newPublishers {

		if limit >= 2 {
			wg.Wait()
		}

		statsLogger("publisher", count, len(newPublishers), v.name)

		limit++
		wg.Add(1)
		go func(publisherID int, v *statsRow) {

			defer func() {
				limit--
				wg.Done()
			}()

			var publisher sql.Publisher

			gorm = gorm.Unscoped().FirstOrInit(&publisher, sql.Publisher{ID: publisherID})
			cronLogErr(gorm.Error)

			publisher.Name = v.name
			publisher.Apps = v.count
			publisher.MeanPrice = v.getMeanPrice()
			publisher.MeanScore = v.getMeanScore()
			publisher.DeletedAt = nil

			gorm = gorm.Unscoped().Save(&publisher)
			cronLogErr(gorm.Error)

		}(k, v)

		count++
	}

	wg.Wait()

	//
	err = sql.SetConfig(sql.ConfPublishersUpdated, strconv.FormatInt(time.Now().Unix(), 10))
	cronLogErr(err)

	page := websockets.GetPage(websockets.PageAdmin) //
	page.Send(websockets.AdminPayload{Message: sql.ConfPublishersUpdated + " complete"})

	//
	err = helpers.GetMemcache().Delete(helpers.MemcachePublisherKeyNames.Key)
	err = helpers.IgnoreErrors(err, helpers.ErrCacheMiss)
	cronLogErr(err)

	//
	cronLogInfo("Publishers updated")
}

func Developers() {

	cronLogInfo("Developers updating")

	// Get current developers, to delete old ones
	allDevelopers, err := sql.GetAllDevelopers([]string{"id", "name"})
	if err != nil {
		cronLogErr(err)
		return
	}

	developersToDelete := map[int]bool{}
	for _, v := range allDevelopers {
		developersToDelete[v.ID] = true
	}

	var developersNameMap = map[int]string{}
	for _, v := range allDevelopers {
		developersNameMap[v.ID] = strings.TrimSpace(v.GetName())
	}

	// Get apps from mysql
	appsWithDevelopers, err := sql.GetAppsWithColumnDepth("developers", 2, []string{"developers", "prices", "reviews_score"})
	cronLogErr(err)

	cronLogInfo("Found " + strconv.Itoa(len(appsWithDevelopers)) + " apps with developers")

	newDevelopers := make(map[int]*statsRow)
	for _, app := range appsWithDevelopers {

		appDevelopers, err := app.GetDeveloperIDs()
		if err != nil {
			cronLogErr(err)
			continue
		}

		if len(appDevelopers) == 0 {
			// appDevelopers = []string{""}
		}

		// For each developer in an app
		for _, appDeveloperID := range appDevelopers {

			delete(developersToDelete, appDeveloperID)

			var developersName string
			if val, ok := developersNameMap[appDeveloperID]; ok {
				developersName = val
			} else {
				continue
			}

			if _, ok := newDevelopers[appDeveloperID]; ok {
				newDevelopers[appDeveloperID].count++
				newDevelopers[appDeveloperID].totalScore += app.ReviewsScore
			} else {
				newDevelopers[appDeveloperID] = &statsRow{
					name:       developersName,
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
				newDevelopers[appDeveloperID].totalPrice[code] += price.Final
			}
		}
	}

	var limit int
	var wg sync.WaitGroup

	// Delete old developers
	limit++
	wg.Add(1)
	go func() {

		defer func() {
			limit--
			wg.Done()
		}()

		var devsToDeleteSlice []int
		for k := range developersToDelete {
			devsToDeleteSlice = append(devsToDeleteSlice, k)
		}

		err := sql.DeleteDevelopers(devsToDeleteSlice)
		cronLogErr(err)

	}()

	wg.Wait()

	gorm, err := sql.GetMySQLClient()
	if err != nil {
		cronLogErr(err)
		return
	}

	// Update current developers
	var count = 1
	for k, v := range newDevelopers {

		if limit >= 2 {
			wg.Wait()
		}

		statsLogger("developer", count, len(newDevelopers), v.name)

		limit++
		wg.Add(1)
		go func(developerInt int, v *statsRow) {

			defer func() {
				limit--
				wg.Done()
			}()

			var developer sql.Developer

			gorm = gorm.Unscoped().FirstOrInit(&developer, sql.Developer{ID: developerInt})
			cronLogErr(gorm.Error)

			developer.Name = v.name
			developer.Apps = v.count
			developer.MeanPrice = v.getMeanPrice()
			developer.MeanScore = v.getMeanScore()
			developer.DeletedAt = nil

			gorm = gorm.Unscoped().Save(&developer)
			cronLogErr(gorm.Error)

		}(k, v)

		count++
	}
	wg.Wait()

	//
	err = sql.SetConfig(sql.ConfDevelopersUpdated, strconv.FormatInt(time.Now().Unix(), 10))
	cronLogErr(err)

	page := websockets.GetPage(websockets.PageAdmin) //
	page.Send(websockets.AdminPayload{Message: sql.ConfDevelopersUpdated + " complete"})

	//
	err = helpers.GetMemcache().Delete(helpers.MemcacheDeveloperKeyNames.Key)
	err = helpers.IgnoreErrors(err, helpers.ErrCacheMiss)
	cronLogErr(err)

	//
	cronLogInfo("Developers updated")
}

func Tags() {

	// Get current tags, to delete old ones
	tags, err := sql.GetAllTags()
	if err != nil {
		cronLogErr(err)
		return
	}

	tagsToDelete := map[int]int{}
	for _, tag := range tags {
		tagsToDelete[tag.ID] = tag.ID
	}

	// Get tags from Steam
	tagsResp, b, err := helpers.GetSteam().GetTags()
	err = helpers.HandleSteamStoreErr(err, b, nil)
	if err != nil {
		cronLogErr(err)
		return
	}

	steamTagMap := tagsResp.GetMap()

	appsWithTags, err := sql.GetAppsWithColumnDepth("tags", 2, []string{"tags", "prices", "reviews_score"})
	cronLogErr(err)

	cronLogInfo("Found " + strconv.Itoa(len(appsWithTags)) + " apps with tags")

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

		defer func() {
			limit--
			wg.Done()
		}()

		var tagsToDeleteSlice []int
		for _, v := range tagsToDelete {
			tagsToDeleteSlice = append(tagsToDeleteSlice, v)
		}

		err := sql.DeleteTags(tagsToDeleteSlice)
		cronLogErr(err)

	}()

	wg.Wait()

	gorm, err := sql.GetMySQLClient()
	if err != nil {
		cronLogErr(err)
		return
	}

	// Update current tags
	var count = 1
	for k, v := range newTags {

		if limit >= 2 {
			wg.Wait()
		}

		statsLogger("tag", count, len(newTags), v.name)

		limit++
		wg.Add(1)
		go func(tagID int, v *statsRow) {

			defer func() {
				limit--
				wg.Done()
			}()

			var tag sql.Tag

			gorm = gorm.Unscoped().FirstOrInit(&tag, sql.Tag{ID: tagID})
			cronLogErr(gorm.Error)

			tag.Name = v.name
			tag.Apps = v.count
			tag.MeanPrice = v.getMeanPrice()
			tag.MeanScore = v.getMeanScore()
			tag.DeletedAt = nil

			gorm = gorm.Unscoped().Save(&tag)
			cronLogErr(gorm.Error)

		}(k, v)

		count++
	}
	wg.Wait()

	//
	err = sql.SetConfig(sql.ConfTagsUpdated, strconv.FormatInt(time.Now().Unix(), 10))
	cronLogErr(err)

	page := websockets.GetPage(websockets.PageAdmin) //
	page.Send(websockets.AdminPayload{Message: sql.ConfTagsUpdated + " complete"})

	//
	err = helpers.GetMemcache().Delete(helpers.MemcacheTagKeyNames.Key)
	err = helpers.IgnoreErrors(err, helpers.ErrCacheMiss)
	cronLogErr(err)

	//
	cronLogInfo("Tags updated")
}

func PlayerRanks() {

	cronLogInfo("Ranks updated started")

	cronLogInfo("Level")
	err := mongo.RankPlayers("level", "level_rank")
	log.Warning(err)

	cronLogInfo("Games")
	err = mongo.RankPlayers("games_count", "games_rank")
	log.Warning(err)

	cronLogInfo("Badges")
	err = mongo.RankPlayers("badges_count", "badges_rank")
	log.Warning(err)

	cronLogInfo("Time")
	err = mongo.RankPlayers("play_time", "play_time_rank")
	log.Warning(err)

	cronLogInfo("Friends")
	err = mongo.RankPlayers("friends_count", "friends_rank")
	log.Warning(err)

	//
	err = sql.SetConfig(sql.ConfRanksUpdated, strconv.FormatInt(time.Now().Unix(), 10))
	cronLogErr(err)

	page := websockets.GetPage(websockets.PageAdmin)
	page.Send(websockets.AdminPayload{Message: sql.ConfRanksUpdated + " complete"})

	cronLogInfo("Ranks updated")
}

func AppPlayers() {

	log.Info("Queueing apps for player checks")

	gorm, err := sql.GetMySQLClient()
	if err != nil {
		log.Critical(err)
		return
	}

	gorm = gorm.Select([]string{"id"})
	gorm = gorm.Order("id ASC")
	gorm = gorm.Model(&[]sql.App{})

	var appIDs []int
	gorm = gorm.Pluck("id", &appIDs)
	if gorm.Error != nil {
		log.Critical(gorm.Error)
	}

	// Chunk appIDs
	var chunks [][]int
	for i := 0; i < len(appIDs); i += 10 {
		end := i + 10

		if end > len(appIDs) {
			end = len(appIDs)
		}

		chunks = append(chunks, appIDs[i:end])
	}

	for _, chunk := range chunks {

		err = queue.ProduceAppPlayers(chunk)
		log.Err(err)
	}

	//
	err = sql.SetConfig(sql.ConfAddedAllAppPlayers, strconv.FormatInt(time.Now().Unix(), 10))
	cronLogErr(err)

	page := websockets.GetPage(websockets.PageAdmin)
	page.Send(websockets.AdminPayload{Message: sql.ConfAddedAllAppPlayers + " complete"})

	cronLogInfo("App players cron complete")
}

func SteamPlayers() {

	log.Info("Cron running: Steam users")

	resp, err := http.Get("https://www.valvesoftware.com/en/about/stats")
	if err != nil {
		log.Err(err)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Err(err)
		return
	}

	sp := steamPlayersStruct{}
	err = helpers.Unmarshal(b, &sp)
	if err != nil {
		log.Err("www.valvesoftware.com/en/about/stats down")
		return
	}

	_, err = helpers.InfluxWrite(helpers.InfluxRetentionPolicyAllTime, influx.Point{
		Measurement: string(helpers.InfluxMeasurementApps),
		Tags: map[string]string{
			"app_id": "0",
		},
		Fields: map[string]interface{}{
			"player_count":  sp.int(sp.InGame),
			"player_online": sp.int(sp.Online),
		},
		Time:      time.Now(),
		Precision: "m",
	})

	log.Warning(err)
}

type steamPlayersStruct struct {
	Online string `json:"users_online"`
	InGame string `json:"users_ingame"`
}

func (sp steamPlayersStruct) int(s string) int {
	s = strings.ReplaceAll(s, ",", "")
	i, err := strconv.Atoi(s)
	log.Warning(err)
	return i
}

func ClearUpcomingCache() {

	var mc = helpers.GetMemcache()
	var err error

	err = mc.Delete(helpers.MemcacheUpcomingAppsCount.Key)
	err = helpers.IgnoreErrors(err, helpers.ErrCacheMiss)
	log.Err(err)

	err = mc.Delete(helpers.MemcacheUpcomingPackagesCount.Key)
	err = helpers.IgnoreErrors(err, helpers.ErrCacheMiss)
	log.Err(err)
}

type statsRow struct {
	name       string
	count      int
	totalPrice map[steam.CountryCode]int
	totalScore float64
}

func (t statsRow) getMeanPrice() string {

	means := map[steam.CountryCode]float64{}

	for code, total := range t.totalPrice {
		means[code] = float64(total) / float64(t.count)
	}

	bytes, err := json.Marshal(means)
	log.Err(err)

	return string(bytes)
}

func (t statsRow) getMeanScore() float64 {
	return t.totalScore / float64(t.count)
}

// Logging
func cronLogErr(interfaces ...interface{}) {
	log.Err(append(interfaces, log.LogNameCron, log.LogNameGameDB)...)
}

func cronLogInfo(interfaces ...interface{}) {
	log.Info(append(interfaces, log.LogNameCron, log.LogNameGameDB)...)
}

func statsLogger(tableName string, count int, total int, rowName string) {

	cronLogInfo("Updating " + tableName + " - " + strconv.Itoa(count) + " / " + strconv.Itoa(total) + ": " + rowName)
}
