package mysql

import (
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/steam-authority/steam-authority/logger"
)

var gormConnection *gorm.DB
var debug = false

func SetDebug(val bool) {
	debug = val
	return
}

func GetDB() (conn *gorm.DB, err error) {

	if gormConnection == nil {

		//logger.Info("Connecting to MySQL")

		db, err := gorm.Open("mysql", os.Getenv("STEAM_SQL_DSN")+"?parseTime=true")
		db.LogMode(debug)
		if err != nil {
			logger.Error(err)
			return db, nil
		}

		gormConnection = db
	}

	return gormConnection, nil
}
