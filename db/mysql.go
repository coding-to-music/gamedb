package db

import (
	"errors"

	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
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

			db, err := gorm.Open("mysql", viper.GetString("MYSQL_DSN")+options)
			if err != nil {
				return db, err
			}
			db.LogMode(true)

			gormConnectionDebug = db
		}

		return gormConnectionDebug, nil
	}

	if gormConnection == nil {

		db, err := gorm.Open("mysql", viper.GetString("MYSQL_DSN")+options)
		if err != nil {
			return db, err
		}

		gormConnection = db
	}

	return gormConnection, nil
}
