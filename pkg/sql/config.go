package sql

import (
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
)

type ConfigType string

func (c ConfigType) String() string {
	return string(c)
}

const (
	// Crons
	ConfTagsUpdated        ConfigType = "refresh-tags"
	ConfGenresUpdated      ConfigType = "refresh-genres"
	ConfClientPlayers      ConfigType = "check-people-on-steam"
	ConfPublishersUpdated  ConfigType = "refresh-publishers"
	ConfDevelopersUpdated  ConfigType = "refresh-developers"
	ConfRanksUpdated       ConfigType = "refresh-ranks"
	ConfAddedAllAppPlayers ConfigType = "refresh-all-app-players"
	ConfClearUpcomingCache ConfigType = "clear-upcoming-cache"
	ConfInstagram          ConfigType = "posted-to-instagram"
	ConfAutoProfile        ConfigType = "auto-profiles-updated"

	//
	ConfAddedAllApps      ConfigType = "refresh-all-apps"
	ConfAddedAllPackages  ConfigType = "refresh-all-packages"
	ConfAddedAllPlayers   ConfigType = "refresh-all-players"
	ConfWipeMemcache      ConfigType = "wipe-memcache"
	ConfRunDevCode        ConfigType = "run-dev-code"
	ConfGarbageCollection ConfigType = "run-garbage-collection"
)

type Config struct {
	ID        string     `gorm:"not null;column:id;primary_key"`
	UpdatedAt *time.Time `gorm:"not null;column:updated_at"`
	Value     string     `gorm:"not null;column:value"`
}

func SetConfig(id ConfigType, value string) (err error) {

	// Update app
	db, err := GetMySQLClient()
	if err != nil {
		return err
	}

	config := &Config{}
	config.ID = string(id)

	db = db.Assign(Config{Value: value}).FirstOrInit(config)
	if db.Error != nil {
		return db.Error
	}

	db = db.Save(config)
	if db.Error != nil {
		return db.Error
	}

	// Save to memcache
	item := helpers.MemcacheConfigItem(string(id))

	return helpers.GetMemcache().SetInterface(item.Key, config, item.Expiration)
}

func GetConfig(id ConfigType) (config Config, err error) {

	var item = helpers.MemcacheConfigItem(string(id))

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

func GetConfigs(ids []ConfigType) (configsMap map[string]Config, err error) {

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
