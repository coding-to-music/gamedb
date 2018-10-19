package db

import (
	"time"
)

const (
	ConfTagsUpdated       = "tags-updated"
	ConfPublishersUpdated = "publishers-updated"
	ConfDevelopersUpdated = "developers-updated"
	ConfGenresUpdated     = "genres-updated"

	ConfRanksUpdated     = "ranks-updated"
	ConfDonationsUpdated = "donations-updated"
	ConfAddedAllApps     = "added-all-apps"
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

	db.Attrs().Assign(Config{Value: value}).FirstOrInit(config)
	db.Save(config)

	return db.Error
}

func GetConfig(id string) (config Config, err error) {

	db, err := GetMySQLClient()
	if err != nil {
		return config, err
	}

	db.Where("id = ?", id).First(&config)

	return config, db.Error
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
