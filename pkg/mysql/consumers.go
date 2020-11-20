package mysql

import (
	"errors"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/ldflags"
	"github.com/gamedb/gamedb/pkg/log"
)

const (
	ConsumerSessionLength  = time.Second * 70 // Time to keep the key for
	consumerSessionRefresh = time.Second * 60 // Heartbeat to retake the key
	consumerSessionRetry   = time.Second * 10 // Retry on no keys availabile
)

type Consumer struct {
	Key         string    `gorm:"not null;column:key;primary_key"`
	Use         bool      `gorm:"not null;column:use;"`
	Expires     time.Time `gorm:"not null;column:expires;type:datetime"`
	Owner       string    `gorm:"not null;column:owner"`
	IP          string    `gorm:"not null;column:ip"`
	Environment string    `gorm:"not null;column:environment"`
	Version     string    `gorm:"not null;column:version"`
	Commits     string    `gorm:"not null;column:commits"`
	Notes       string    `gorm:"-"`
}

func GetConsumer(tag string) (err error) {

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
		var row = Consumer{}

		db = db.Where("`expires` < ?", time.Now())
		db = db.Where("`use` = ?", 1)
		db = db.Set("gorm:query_option", "FOR UPDATE") // Locks row
		db = db.Order("`expires` ASC")
		db = db.First(&row)

		if db.Error == ErrRecordNotFound {
			db.Rollback()
			return errors.New("waiting for consumer")
		} else if db.Error != nil {
			db.Rollback()
			return db.Error
		}

		if row.Key == "" {
			db.Rollback()
			return errors.New("invalid empty key")
		}

		// Update the row
		fields := map[string]interface{}{
			"expires":     time.Now().Add(ConsumerSessionLength),
			"owner":       tag,
			"environment": config.C.Environment,
			"version":     config.GetShortCommitHash(),
			"ip":          config.C.IP,
			"commits":     ldflags.CommitCount,
		}

		db = db.New().Table("consumers").Where("`key` = ?", row.Key).Updates(fields)
		if db.Error != nil {
			db.Rollback()
			return db.Error
		}

		//
		db = db.Commit()

		if db.Error == nil {
			config.C.SteamAPIKey = row.Key
		}

		return db.Error
	}

	policy := backoff.NewConstantBackOff(consumerSessionRetry)
	err = backoff.RetryNotify(operation, policy, func(err error, t time.Duration) { log.InfoS(err) })
	if err != nil {
		return err
	}

	// Keep the key in use with a heartbeat
	go func() {

		db, err := GetMySQLClient()
		if err != nil {
			log.ErrS(err)
			return
		}

		db = db.Model(&Consumer{})
		db = db.Where("`key` = ?", config.C.SteamAPIKey)

		for {

			time.Sleep(consumerSessionRefresh)

			// Update key
			fields := map[string]interface{}{
				"expires": time.Now().Add(ConsumerSessionLength),
			}

			db = db.Updates(fields)
			if db.Error != nil {
				log.ErrS(db.Error)
			}
		}
	}()

	return err
}
