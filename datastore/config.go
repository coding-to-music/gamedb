package datastore

import (
	"strings"
	"time"

	"cloud.google.com/go/datastore"
)

const (
	ConfTagsUpdated       = "tags-updated"
	ConfPublishersUpdated = "publishers-updated"
	ConfDevelopersUpdated = "developers-updated"
	ConfRanksUpdated      = "ranks-updated"
	ConfGenresUpdated     = "genres-updated"
	ConfDonationsUpdated  = "donations-updated"
	ConfDeployed          = "deployed"
	ConfAddedAllApps      = "added-all-apps"
)

type Config struct {
	UpdatedAt time.Time `datastore:"updated_at,noindex"`
	ConfigID  string    `datastore:"config_id"`
	Value     string    `datastore:"value,noindex"`
}

func (c Config) GetKey() (key *datastore.Key) {
	return datastore.NameKey(KindConfig, c.ConfigID, nil)
}

func SetConfig(configID string, value string) (err error) {

	config := new(Config)
	config.UpdatedAt = time.Now()
	config.ConfigID = configID
	config.Value = value

	_, err = SaveKind(config.GetKey(), config)

	return err
}

func GetConfig(name string) (config *Config, err error) {

	client, context, err := getClient()
	if err != nil {
		return config, err
	}
	key := datastore.NameKey(KindConfig, name, nil)

	config = new(Config)

	err = client.Get(context, key, config)
	if err != nil {

		if err.Error() == ErrorNotFound {
			config.ConfigID = name
			return config, nil
		}
		return config, err
	}
	return config, nil
}

func GetMultiConfigs(names []string) (configsMap map[string]Config, err error) {

	configsMap = map[string]Config{}

	client, context, err := getClient()
	if err != nil {
		return configsMap, err
	}

	var keys []*datastore.Key
	for _, v := range names {
		key := datastore.NameKey(KindConfig, v, nil)
		keys = append(keys, key)
	}

	configs := make([]Config, len(keys))
	err = client.GetMulti(context, keys, configs)
	if err != nil && !strings.Contains(err.Error(), "no such entity") {
		return configsMap, err
	}

	// Make into map
	for _, v := range configs {
		if v.ConfigID != "" {
			configsMap[v.ConfigID] = v
		}
	}

	return configsMap, nil
}
