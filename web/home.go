package web

import (
	"net/http"
	"sync"

	"github.com/gamedb/website/db"
	"github.com/gamedb/website/log"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {

	t := homeTemplate{}
	t.Fill(w, r, "Home", "Stats and information on the Steam Catalogue.")

	var wg sync.WaitGroup

	// wg.Add(1)
	// go func() {
	//
	// 	defer wg.Done()
	//
	// 	var err error
	// 	t.RanksCount, err = db.CountRanks()
	// 	log.Err(err, r)
	// }()
	//
	// wg.Add(1)
	// go func() {
	//
	// 	defer wg.Done()
	//
	// 	var err error
	// 	t.AppsCount, err = db.CountApps()
	// 	log.Err(err, r)
	// }()
	//
	// wg.Add(1)
	// go func() {
	//
	// 	defer wg.Done()
	//
	// 	var err error
	// 	t.PackagesCount, err = db.CountPackages()
	// 	log.Err(err, r)
	// }()

	wg.Add(1)
	go func() {

		defer wg.Done()

		gorm, err := db.GetMySQLClient()
		if err != nil {
			log.Err(err)
			return
		}

		gorm = gorm.Select([]string{"id", "name", "icon", "player_count"})
		gorm = gorm.Order("player_count desc")
		gorm = gorm.Limit(15)
		gorm = gorm.Find(&t.PopularApps)

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
		gorm = gorm.Find(&t.RatedApps)

		log.Err(err, r)
	}()

	wg.Wait()

	err := returnTemplate(w, r, "home", t)
	log.Err(err, r)
}

type homeTemplate struct {
	GlobalTemplate
	RanksCount    int
	AppsCount     int
	PackagesCount int
	PopularApps   []db.App
	RatedApps     []db.App
}
