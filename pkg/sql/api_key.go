package sql

import (
	"errors"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
)

const (
	apiSessionLength  = time.Second * 70 // Time to keep the key for
	apiSessionRefresh = time.Second * 60 // Heartbeat to retake the key
	apiSessionRetry   = time.Second * 10 // Retry on no keys availabile
)

type APIKey struct {
	Key     string    `gorm:"not null;column:key;PRIMARY_KEY"`
	Use     bool      `gorm:"not null;column:use;"`
	Expires time.Time `gorm:"not null;column:expires;type:datetime"`
	Owner   string    `gorm:"not null;column:owner"`
	IP      string    `gorm:"not null;column:ip"`
	Notes   string    `gorm:"-"`
}

type Lock struct {
	Success bool `gorm:"not null;column:success"`
}

func GetAPIKey(tag string) (err error) {

	// Retrying as this call can fail
	operation := func() (err error) {

		db, err := GetMySQLClient()
		if err != nil {
			return err
		}

		db = db.Begin()
		if db.Error != nil {
			return db.Error
		}

		// Find a row
		var row = APIKey{}

		db = db.Where("`expires` < ?", time.Now())
		db = db.Where("`use` = ?", 1)
		db = db.Set("gorm:query_option", "FOR UPDATE") // Locks row
		db = db.Order("`expires` ASC")
		db = db.First(&row)

		if db.Error == ErrRecordNotFound {
			db.Rollback()
			return errors.New("waiting for API key")
		} else if db.Error != nil {
			db.Rollback()
			return db.Error
		}

		// Update the row
		db = db.New().Table("api_keys").Where("`key` = ?", row.Key).Updates(map[string]interface{}{
			"expires":     time.Now().Add(apiSessionLength),
			"owner":       tag,
			"environment": config.Config.Environment.Get(),
			"version":     config.GetShortCommitHash(),
			"ip":          config.Config.IP.Get(),
		})
		if db.Error != nil {
			db.Rollback()
			return db.Error
		}

		//
		db = db.Commit()

		if db.Error == nil {
			config.Config.SteamAPIKey.SetDefault(row.Key)
			log.Info("Using Steam API key: " + config.GetSteamKeyTag())
		}

		return db.Error
	}

	policy := backoff.NewConstantBackOff(apiSessionRetry)
	err = backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { log.Info(err) })
	if err != nil {
		return err
	}

	// Keep the key in use with a heartbeat
	go func() {

		db, err := GetMySQLClient()
		if err != nil {
			log.Err(err)
			return
		}

		db = db.Model(&APIKey{})
		db = db.Where("`key` = ?", config.Config.SteamAPIKey.Get())

		for {
			time.Sleep(apiSessionRefresh)

			// Update key
			db = db.Updates(map[string]interface{}{
				"expires": time.Now().Add(apiSessionLength),
			})
			if db.Error != nil {
				log.Err(db.Error)
			}
		}
	}()

	return err
}
