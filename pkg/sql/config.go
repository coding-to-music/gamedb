package sql

import (
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
)

type ConfigType string

func (c ConfigType) String() string {
	return string(c)
}

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

func GetAllConfigs() (configsMap map[string]Config, err error) {

	configsMap = map[string]Config{}

	db, err := GetMySQLClient()
	if err != nil {
		return configsMap, err
	}

	var configs []Config
	db.Find(&configs)
	if db.Error != nil {
		return configsMap, db.Error
	}

	for _, v := range configs {
		configsMap[v.ID] = v
	}

	return configsMap, nil
}
