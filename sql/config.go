package sql

import (
	"time"

	"github.com/gamedb/website/helpers"
)

const (
	ConfTagsUpdated       = "refresh-tags"
	ConfGenresUpdated     = "refresh-genres"
	ConfPublishersUpdated = "refresh-publishers"
	ConfDevelopersUpdated = "refresh-developers"

	ConfRanksUpdated      = "refresh-ranks"
	ConfDonationsUpdated  = "refresh-donations"
	ConfAddedAllApps      = "refresh-all-apps"
	ConfAddedAllPackages  = "refresh-all-packages"
	ConfAddedAllPlayers   = "refresh-all-players"
	ConfWipeMemcache      = "wipe-memcache"
	ConfRunDevCode        = "run-dev-code"
	ConfGarbageCollection = "run-garbage-collection"
)

type Config struct {
	ID        string     `gorm:"not null;column:id;primary_key"`
	UpdatedAt *time.Time `gorm:"not null;column:updated_at"`
	Value     string     `gorm:"not null;column:value"`
}

func SetConfig(id string, value string) (err error) {

	// Update app
	db, err := GetMySQLClient()
	if err != nil {
		return err
	}

	config := new(Config)
	config.ID = id

	db = db.Assign(Config{Value: value}).FirstOrInit(config)
	if db.Error != nil {
		return db.Error
	}

	db = db.Save(config)
	if db.Error != nil {
		return db.Error
	}

	// Save to memcache
	item := helpers.MemcacheConfigRow(id)

	return helpers.GetMemcache().SetInterface(item.Key, config, item.Expiration)
}

func GetConfig(id string) (config Config, err error) {

	var item = helpers.MemcacheConfigRow(id)

	err = helpers.GetMemcache().GetSetInterface(item.Key, item.Expiration, &config, func() (interface{}, error) {

		var config Config

		db, err := GetMySQLClient()
		if err != nil {
			return config, err
		}

		db = db.Where("id = ?", id).First(&config)

		return config, db.Error
	})

	return config, err
}

func GetConfigs(ids []string) (configsMap map[string]Config, err error) {

	configsMap = map[string]Config{}

	if len(ids) == 0 {
		return configsMap, nil
	}

	db, err := GetMySQLClient()
	if err != nil {
		return configsMap, err
	}

	var configs []Config
	db.Where("id IN (?)", ids).Find(&configs)
	if db.Error != nil {
		return configsMap, db.Error
	}

	for _, v := range configs {
		configsMap[v.ID] = v
	}

	return configsMap, nil
}
