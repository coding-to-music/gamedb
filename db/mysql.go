package db

import (
	"net/url"
	"sync"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/gamedb/website/config"
	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/jinzhu/gorm"
)

var (
	ErrRecordNotFound = gorm.ErrRecordNotFound

	gormConnection      *gorm.DB
	gormConnectionDebug *gorm.DB

	SQLMutex sync.Mutex
)

func GetMySQLClient(debug ...bool) (conn *gorm.DB, err error) {

	SQLMutex.Lock()
	defer SQLMutex.Unlock()

	// Retrying as this call can fail
	operation := func() (err error) {

		if len(debug) == 0 {
			if gormConnection == nil {
				gormConnection, err = getMySQLConnection()
				if err != nil {
					return err
				}
				gormConnection.SetLogger(mySQLLogger{})
			}
			conn = gormConnection
		} else {
			if gormConnectionDebug == nil {
				gormConnectionDebug, err = getMySQLConnection()
				if err != nil {
					return err
				}
				gormConnectionDebug.SetLogger(mySQLLoggerDebug{})
			}
			conn = gormConnectionDebug
		}

		return pingMySQL(gormConnection)
	}

	policy := backoff.NewExponentialBackOff()

	err = backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { log.Info(err) })
	if err != nil {
		log.Critical(err)
	}

	return conn, err
}

func getMySQLConnection() (*gorm.DB, error) {

	log.Info("Connecting to MySQL")

	options := url.Values{}
	options.Set("parseTime", "true")
	options.Set("charset", "utf8mb4")
	options.Set("collation", "utf8mb4_unicode_ci")

	db, err := gorm.Open("mysql", config.Config.MySQLDNS()+"?"+options.Encode())
	if err != nil {

		return nil, err
	}
	db = db.LogMode(true)
	db = db.Set("gorm:association_autoupdate", false)
	db = db.Set("gorm:association_autocreate", false)
	db = db.Set("gorm:association_save_reference", false)
	db = db.Set("gorm:save_associations", false)

	return db, err
}

func pingMySQL(gorm *gorm.DB) error {
	gorm = gorm.Exec("SELECT VERSION()")
	return gorm.Error
}

type mySQLLogger struct {
}

func (logger mySQLLogger) Print(v ...interface{}) {

	s := helpers.JoinInterface(v)
	log.Debug(s, log.LogNameSQL, log.ServiceGoogle)
}

type mySQLLoggerDebug struct {
}

func (logger mySQLLoggerDebug) Print(v ...interface{}) {

	s := helpers.JoinInterface(v)
	log.Debug(s, log.LogNameSQL, log.ServiceLocal)
}
