package db

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"errors"
	"sort"
	"strings"

	"github.com/Masterminds/squirrel"
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
func Select(builder squirrel.SelectBuilder) (rows *sql.Rows, err error) {

	rawSQL, args, err := builder.ToSql()
	if err != nil {
		return rows, err
	}

	prep, err := getPrepareStatement(rawSQL)
	if err != nil {
		return rows, err
	}

	rows, err = prep.Query(args...)
	if err != nil {
		return rows, err
	}

	return rows, nil
}

func SelectFirst(builder squirrel.SelectBuilder) (row *sql.Row, err error) {

	builder.Limit(1)

	rawSQL, args, err := builder.ToSql()
	if err != nil {
		return row, err
	}

	prep, err := getPrepareStatement(rawSQL)
	if err != nil {
		return row, err
	}

	return prep.QueryRow(args...), nil
}

func Insert(builder squirrel.InsertBuilder) (result sql.Result, err error) {

	rawSQL, args, err := builder.ToSql()
	if err != nil {
		return result, err
	}

	prep, err := getPrepareStatement(rawSQL)
	if err != nil {
		return result, err
	}

	result, err = prep.Exec(args...)
	if err != nil {
		return result, err
	}

	return result, nil
}

func Update(builder squirrel.UpdateBuilder) (result sql.Result, err error) {

	rawSQL, args, err := builder.ToSql()
	if err != nil {
		return result, err
	}

	prep, err := getPrepareStatement(rawSQL)
	if err != nil {
		return result, err
	}

	result, err = prep.Exec(args...)
	if err != nil {
		return result, err
	}

	return result, nil
}

func RawQuery(query string, args []interface{}) (result sql.Result, err error) {

	prep, err := getPrepareStatement(query)
	if err != nil {
		return result, err
	}

	result, err = prep.Exec(args...)
	if err != nil {
		return result, err
	}

	return result, nil
}

func UpdateInsert(table string, data UpdateInsertData) (result sql.Result, err error) {

	query := "INSERT INTO " + table + " (" + data.formattedColumns() + ") VALUES (" + data.getMarks() + ") ON DUPLICATE KEY UPDATE " + data.getDupes() + ";"
	return RawQuery(query, data.getValues())
}

var mysqlPrepareStatements map[string]*sql.Stmt

func getPrepareStatement(query string) (statement *sql.Stmt, err error) {

	if mysqlPrepareStatements == nil {
		mysqlPrepareStatements = make(map[string]*sql.Stmt)
	}

	byteArray := md5.Sum([]byte(query))
	hash := hex.EncodeToString(byteArray[:])

	if val, ok := mysqlPrepareStatements[hash]; ok {
		if ok {
			return val, nil
		}
	}

	conn, err := getMysqlConnection()
	if err != nil {
		return statement, err
	}

	statement, err = conn.Prepare(query)
	if err != nil {
		return statement, err
	}

	mysqlPrepareStatements[hash] = statement
	return statement, nil
}

var mysqlConnection *sql.DB

func getMysqlConnection() (db *sql.DB, err error) {

	if mysqlConnection == nil {

		var err error
		mysqlConnection, err = sql.Open("mysql", viper.GetString("MYSQL_DSN"))
		if err != nil {
			return db, err
		}
	}

	return mysqlConnection, nil
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
