package db

import (
	"errors"
	"sort"
	"strings"

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

//
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

//
func UpdateInsert(table string, data UpdateInsertData) error {

	db, err := GetMySQLClient()
	if err != nil {
		return err
	}

	query := "INSERT INTO " + table + " (" + data.formattedColumns() + ") VALUES (" + data.getMarks() + ") ON DUPLICATE KEY UPDATE " + data.getDupes() + ";"
	return db.Exec(query).Error
}

// UpdateInsertData
type UpdateInsertData map[string]interface{}

func (ui UpdateInsertData) sortedColumns() (columns []string) {

	var slice []string
	for k := range ui {
		slice = append(slice, k)
	}
	sort.Strings(slice)
	return slice
}

func (ui UpdateInsertData) formattedColumns() (columns string) {

	var slice []string
	for _, v := range ui.sortedColumns() {
		slice = append(slice, "`"+v+"`")
	}
	return strings.Join(slice, ", ")
}

func (ui UpdateInsertData) getDupes() (columns string) {

	var slice []string
	for _, v := range ui.sortedColumns() {
		slice = append(slice, v+"=VALUES("+v+")")
	}
	return strings.Join(slice, ", ")
}

func (ui UpdateInsertData) getValues() (columns []interface{}) {

	var slice []interface{}
	for _, v := range ui.sortedColumns() {
		slice = append(slice, ui[v])
	}
	return slice
}

func (ui UpdateInsertData) getMarks() (marks string) {

	var slice []string
	for range ui {
		slice = append(slice, "?")
	}
	return strings.Join(slice, ", ")
}
