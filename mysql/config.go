package mysql

import (
	"time"

	"github.com/steam-authority/steam-authority/logger"
)

type Config struct {
	ID        string     `gorm:"not null;column:id;primary_key"`
	UpdatedAt *time.Time `gorm:"not null;column:updated_at"`
	Value     string     `gorm:"not null;column:value"`
}

func SetConfig(id string, value string) (err error) {

	// Update app
	db, err := GetDB()
	if err != nil {
		logger.Error(err)
	}

	config := new(Config)
	config.ID = id

	db.Attrs().Assign(Config{Value: value}).FirstOrInit(config)

	db.Save(config)
	if db.Error != nil {
		logger.Error(err)
	}

	return nil
}

func GetConfig(id string) (config Config, err error) {

	db, err := GetDB()
	if err != nil {
		return config, err
	}

	db.First(&config, id)
	if db.Error != nil {
		return config, db.Error
	}

	return config, nil
}

func GetConfigs(ids []string) (configsMap map[string]Config, err error) {

	configsMap = map[string]Config{}

	if len(ids) < 1 {
		return configsMap, nil
	}

	db, err := GetDB()
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
