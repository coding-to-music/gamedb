package mysql

import (
	"net/url"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/gamedb/gamedb/pkg/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
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

			// zap.S().Info("Connecting to MySQL")

			options := url.Values{}
			options.Set("parseTime", "true")
			options.Set("charset", "utf8mb4")
			options.Set("collation", "utf8mb4_unicode_ci")

			conn, err := gorm.Open("mysql", config.MySQLDNS()+"?"+options.Encode())
			if err != nil {
				return err
			}
			conn = conn.Set("gorm:association_autoupdate", false)
			conn = conn.Set("gorm:association_autocreate", false)
			conn = conn.Set("gorm:association_save_reference", false)
			conn = conn.Set("gorm:save_associations", false)
			conn = conn.LogMode(false)
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

		err = backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { zap.S().Info(err) })
		if err != nil {
			zap.S().Fatal(err)
		}
	}

	return gormConnection, err
}

type mySQLLogger struct {
}

func (logger mySQLLogger) Print(v ...interface{}) {
	zap.S().Debug(v...)
}
