package influx

import (
	"net/url"
	"sync"

	"github.com/gamedb/gamedb/pkg/config"
	influx "github.com/influxdata/influxdb1-client"
)

var (
	client *influx.Client
	lock   sync.Mutex
)

func getInfluxClient() (*influx.Client, error) {

	lock.Lock()
	defer lock.Unlock()

	var err error
	var host *url.URL

	if client == nil {

		host, err = url.Parse(config.Config.InfluxURL.Get())
		if err != nil {
			return nil, err
		}

		client, err = influx.NewClient(influx.Config{
			URL:      *host,
			Username: config.Config.InfluxUsername.Get(),
			Password: config.Config.InfluxPassword.Get(),
		})
	}

	return client, err
}
