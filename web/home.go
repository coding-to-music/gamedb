package web

import (
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/gamedb/website/cache"
	"github.com/gamedb/website/db"
	"github.com/gamedb/website/log"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {

	t := homeTemplate{}
	t.Fill(w, r, "Home", "Stats and information on the Steam Catalogue.")

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.PlayersCount, err = db.CountPlayers()
		log.Err(err, r)
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.AppsCount, err = db.CountApps()
		log.Err(err, r)
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.PackagesCount, err = db.CountPackages()
		log.Err(err, r)
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error
		t.PopularApps, err = cache.PopularApps()
		log.Err(err, r)
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		gorm, err := db.GetMySQLClient()
		if err != nil {
			log.Err(err)
			return
		}

		gorm = gorm.Select([]string{"id", "name", "icon", "reviews_score"})
		gorm = gorm.Where("type = ?", "game")
		gorm = gorm.Order("reviews_score desc")
		gorm = gorm.Limit(15)
		gorm = gorm.Find(&t.TrendingApps)

		log.Err(err, r)
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		gorm, err := db.GetMySQLClient()
		if err != nil {
			log.Err(err)
			return
		}

		gorm = gorm.Select([]string{"id", "name", "icon", "player_peak_week"})
		gorm = gorm.Where("type = ?", "game")
		gorm = gorm.Where("release_date_unix > ?", time.Now().Add(time.Hour * 24 * 7 * -1).Unix())
		gorm = gorm.Order("player_peak_week desc")
		gorm = gorm.Limit(15)
		gorm = gorm.Find(&t.PopularNewApps)

		log.Err(err, r)
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		gorm, err := db.GetMySQLClient()
		if err != nil {
			log.Err(err)
			return
		}

		gorm = gorm.Select([]string{"id", "name", "icon", "reviews_score"})
		gorm = gorm.Where("type = ?", "game")
		gorm = gorm.Where("release_date_unix > ?", time.Now().Add(time.Hour * 24 * 7 * -1).Unix())
		gorm = gorm.Order("reviews_score desc")
		gorm = gorm.Limit(15)
		gorm = gorm.Find(&t.RatedNewApps)

		log.Err(err, r)
	}()

	wg.Wait()

	// Get news
	client, ctx, err := db.GetDSClient()
	if err != nil {

		log.Err(err, r)
		return
	}

	var allNews []db.News
	for _, v := range t.PopularApps {

		var news []db.News
		q := datastore.NewQuery(db.KindNews).Filter("app_id =", v.ID).Order("-date").Limit(3)
		_, err = client.GetAll(ctx, q, &news)
		err = db.HandleDSMultiError(err, db.OldNewsFields)
		log.Err(err, r)
		allNews = append(allNews, news...)
		fmt.Println(len(news))
	}

	sort.Slice(allNews, func(i, j int) bool {
		return allNews[i].Date.Unix() < allNews[j].Date.Unix()
	})

	t.News = allNews

	//
	err = returnTemplate(w, r, "home", t)
	log.Err(err, r)
}

type homeTemplate struct {
	GlobalTemplate
	AppsCount      int
	PackagesCount  int
	PlayersCount   int
	PopularApps    []db.App
	TrendingApps   []db.App
	RatedNewApps   []db.App
	PopularNewApps []db.App
	News           []db.News
}
