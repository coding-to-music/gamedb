package sql

import (
	"net/url"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

var (
	ErrRecordNotFound = gorm.ErrRecordNotFound

	gormConnection      *gorm.DB
	gormConnectionMutex sync.Mutex
)

func GetMySQLClient() (conn *gorm.DB, err error) {

	gormConnectionMutex.Lock()
	defer gormConnectionMutex.Unlock()

	if gormConnection == nil {

		// Retrying as this call can fail
		operation := func() (err error) {

			// log.Info("Connecting to MySQL")

			options := url.Values{}
			options.Set("parseTime", "true")
			options.Set("charset", "utf8mb4")
			options.Set("collation", "utf8mb4_unicode_ci")

			conn, err := gorm.Open("mysql", config.MySQLDNS()+"?"+options.Encode())
			if err != nil {
				return err
			}
			conn = conn.LogMode(false)
			conn = conn.Set("gorm:association_autoupdate", false)
			conn = conn.Set("gorm:association_autocreate", false)
			conn = conn.Set("gorm:association_save_reference", false)
			conn = conn.Set("gorm:save_associations", false)
			conn.SetLogger(mySQLLogger{})

			// test ping
			conn = conn.Exec("SELECT VERSION()")
			if conn.Error != nil {
				return conn.Error
			}

			gormConnection = conn

			return err
		}

		policy := backoff.NewConstantBackOff(time.Second * 5)

		err = backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { log.Info(err) })
		if err != nil {
			log.Critical(err)
		}
	}

	return gormConnection, err
}

type mySQLLogger struct {
}

func (logger mySQLLogger) Print(v ...interface{}) {
	s := helpers.InterfaceToString(v)
	log.Debug(s, log.LogNameSQL)
}
