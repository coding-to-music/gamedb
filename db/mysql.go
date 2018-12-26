package db

import (
	"errors"

	"github.com/gamedb/website/config"
	"github.com/jinzhu/gorm"
)

var (
	ErrNotFound = errors.New("not found")

	gormConnection      *gorm.DB
	gormConnectionDebug *gorm.DB
)

func GetMySQLClient(debug ...bool) (conn *gorm.DB, err error) {

	var options = "?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci"

	if len(debug) > 0 {

		if gormConnectionDebug == nil {

			db, err := gorm.Open("mysql", config.Config.MySQLDSN.Get()+options)
			if err != nil {
				return db, err
			}
			db.LogMode(true)

			gormConnectionDebug = db
		}

		return gormConnectionDebug, nil
	}

	if gormConnection == nil {

		db, err := gorm.Open("mysql", config.Config.MySQLDSN.Get()+options)
		if err != nil {
			return db, err
		}

		gormConnection = db
	}

	return gormConnection, nil
}
