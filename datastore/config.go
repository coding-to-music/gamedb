package datastore

import (
	"time"

	"cloud.google.com/go/datastore"
)

type Config struct {
	CreatedAt time.Time `datastore:"created_at,noindex"`
	UpdatedAt time.Time `datastore:"updated_at,noindex"`
	ConfigID  string    `datastore:"config_id"`
	Value     string    `datastore:"apps,noindex"`
}

func (c Config) GetKey() (key *datastore.Key) {
	return datastore.IncompleteKey(KindConfig, nil)
}

func SetConfig(name string, value string) (err error) {

	config, err := GetConfig(name)
	if err != nil {
		return err
	}

	if config.CreatedAt.IsZero() {
		config.CreatedAt = time.Now()
	}
	config.UpdatedAt = time.Now()

	_, err = SaveKind(config.GetKey(), config)

	return err
}

func GetConfig(name string) (config *Config, err error) {

	client, context, err := getDSClient()
	if err != nil {
		return config, err
	}
	key := datastore.NameKey(KindConfig, name, nil)

	config = new(Config)

	err = client.Get(context, key, config)
	if err != nil {

		if err.Error() == ErrorNotFound {
			return config, nil
		}
		return config, err
	}
	return config, nil
}
