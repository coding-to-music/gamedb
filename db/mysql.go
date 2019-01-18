package db

import (
	"net/url"
	"sync"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/gamedb/website/config"
	"github.com/gamedb/website/log"
	"github.com/jinzhu/gorm"
)

var (
	ErrRecordNotFound = gorm.ErrRecordNotFound

	gormConnection      *gorm.DB
	gormConnectionDebug *gorm.DB

	SQLMutex = new(sync.Mutex)
)

func GetMySQLClient(debug ...bool) (conn *gorm.DB, err error) {

	SQLMutex.Lock()
	defer SQLMutex.Unlock()

	// Retrying as this call can fail
	operation := func() (err error) {

		options := url.Values{}
		options.Set("parseTime", "true")
		options.Set("charset", "utf8mb4")
		options.Set("collation", "utf8mb4_unicode_ci")

		if len(debug) > 0 {

			if gormConnectionDebug == nil {

				log.Info("Connecting to MySQL")

				db, err := gorm.Open("mysql", config.Config.MySQLDSN.Get()+"?"+options.Encode())
				if err != nil {
					return err
				}
				db.LogMode(true)
				db.SetLogger(MySQLLogger{})

				gormConnectionDebug = db
			}

			conn = gormConnectionDebug
			return nil
		}

		if gormConnection == nil {

			log.Info("Connecting to MySQL")

			db, err := gorm.Open("mysql", config.Config.MySQLDSN.Get()+"?"+options.Encode())
			if err != nil {
				return err
			}
			db.LogMode(true)
			db.SetLogger(MySQLLogger{})

			gormConnection = db
		}

		conn = gormConnection
		return nil
	}

	policy := backoff.NewExponentialBackOff()

	err = backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { log.Info(err) })
	if err != nil {
		log.Critical(err)
	}

	return conn, err
}

type MySQLLogger struct {
}

func (ml MySQLLogger) Print(v ...interface{}) {
	log.Debug(append(v, log.LogNameSQL, log.EnvProd)...)
}
