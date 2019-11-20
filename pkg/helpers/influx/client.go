package influx

import (
	"net/url"
	"sync"

	"github.com/gamedb/gamedb/pkg/config"
	influx "github.com/influxdata/influxdb1-client"
)

var (
	globalClient *influx.Client
	lock         sync.Mutex
)

func getInfluxClient() (*influx.Client, error) {

	lock.Lock()
	defer lock.Unlock()

	var err error

	if globalClient == nil {

		host, err := url.Parse(config.Config.InfluxURL.Get())
		if err != nil {
			return nil, err
		}

		conf := influx.Config{
			URL:      *host,
			Username: config.Config.InfluxUsername.Get(),
			Password: config.Config.InfluxPassword.Get(),
		}

		globalClient, err = influx.NewClient(conf)
	}

	return globalClient, err
}
