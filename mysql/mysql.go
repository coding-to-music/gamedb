package mysql

import (
	"errors"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var (
	gormConnection *gorm.DB
	debug          = false

	ErrNotFound = errors.New("not found")
)

func SetDebug(val bool) {
	debug = val
	return
}

func GetDB() (conn *gorm.DB, err error) {

	if gormConnection == nil {

		username := os.Getenv("STEAM_MYSQL_USERNAME")
		password := os.Getenv("STEAM_MYSQL_PASSWORD")
		database := os.Getenv("STEAM_MYSQL_DATABASE")
		host := os.Getenv("STEAM_MYSQL_HOST")
		port := os.Getenv("STEAM_MYSQL_PORT")

		db, err := gorm.Open("mysql", username+":"+password+"@tcp("+host+":"+port+")/"+database+"?parseTime=true")
		db.LogMode(debug)
		if err != nil {
			return db, nil
		}

		gormConnection = db
	}

	return gormConnection, nil
}

type UpdateError struct {
	err  string
	hard bool
	log  bool
}

func (e UpdateError) Error() string {
	return e.err
}

func (e UpdateError) IsHard() bool {
	return e.hard
}

func (e UpdateError) IsSoft() bool {
	return !e.hard
}

func (e UpdateError) Log() bool {
	return e.log
}
