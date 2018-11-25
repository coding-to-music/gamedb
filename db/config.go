package db

import (
	"encoding/json"
	"time"

	"github.com/gamedb/website/helpers"
)

const (
	ConfTagsUpdated       = "refresh-tags"
	ConfGenresUpdated     = "refresh-genres"
	ConfPublishersUpdated = "refresh-publishers"
	ConfDevelopersUpdated = "refresh-developers"

	ConfRanksUpdated     = "refresh-ranks"
	ConfDonationsUpdated = "refresh-donations"
	ConfAddedAllApps     = "refresh-all-apps"
	ConfWipeMemcache     = "wipe-memcache"
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

	db = db.Attrs().Assign(Config{Value: value}).FirstOrInit(config)
	db = db.Save(config)
	if db.Error != nil {
		return db.Error
	}

	// Save to memcache
	item := helpers.MemcacheConfigRow(id)
	return helpers.GetMemcache().Set(item.Key, value, item.Expiration)
}

func GetConfig(id string) (config Config, err error) {

	s, err := helpers.GetMemcache().GetSetString(helpers.MemcacheConfigRow(id), func() (s string, err error) {

		db, err := GetMySQLClient()
		if err != nil {
			return s, err
		}

		db.Where("id = ?", id).First(&config)
		if db.Error != nil {
			return s, db.Error
		}

		bytes, err := json.Marshal(config)
		return string(bytes), err
	})

	if err != nil {
		return config, err
	}

	err = helpers.Unmarshal([]byte(s), &config)
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
