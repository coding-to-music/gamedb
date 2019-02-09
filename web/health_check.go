package web

import (
	"net/http"

	"github.com/gamedb/website/db"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
)

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/plain")

	code := http.StatusOK

	// Check MySQL
	gorm, err := db.GetMySQLClient()
	if err != nil {
		gorm = gorm.Exec("SELECT version()")
		if gorm.Error != nil {
			log.Err(gorm.Error, r)
			code = http.StatusInternalServerError
		}
	}

	// Check
	var i int
	err = helpers.GetMemcache().Get(helpers.MemcacheAppsCount.Key, &i)
	if err != nil {
		log.Err(err, r)
		code = http.StatusInternalServerError
	}

	w.WriteHeader(code)

	if code == http.StatusOK {
		_, err = w.Write([]byte("OK"))
	} else {
		_, err = w.Write([]byte("ERROR"))
	}
	log.Err(err, r)
}
