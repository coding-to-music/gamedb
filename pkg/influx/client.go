package influx

import (
	"fmt"
	"net/url"
	"sync"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	influx "github.com/influxdata/influxdb1-client"
	"go.uber.org/zap"
)

var (
	client *influx.Client
	mutex  sync.Mutex
)

func getInfluxClient() (*influx.Client, error) {

	mutex.Lock()
	defer mutex.Unlock()

	var err error
	var host *url.URL

	if client == nil {

		host, err = url.Parse(config.C.InfluxURL)
		if err != nil {
			return nil, err
		}

		client, err = influx.NewClient(influx.Config{
			URL:      *host,
			Username: config.C.InfluxUsername,
			Password: config.C.InfluxPassword,
		})
	}

	return client, err
}

var (
	client2 influxdb2.Client
	mutex2  sync.Mutex
)

func getInfluxClient2() influxdb2.Client {

	mutex2.Lock()
	defer mutex2.Unlock()

	if client2 == nil {
		client2 = influxdb2.NewClient(config.C.InfluxURL, fmt.Sprintf("%s:%s", config.C.InfluxUsername, config.C.InfluxPassword))
	}

	return client2
}

var (
	reader      api.QueryAPI
	readerMutex sync.Mutex
)

func getReader() api.QueryAPI {

	readerMutex.Lock()
	defer readerMutex.Unlock()

	if reader == nil {
		reader = getInfluxClient2().QueryAPI("")
	}

	return reader
}

var (
	writer      api.WriteAPI
	writerMutex sync.Mutex
)

func GetWriter() api.WriteAPI {

	writerMutex.Lock()
	defer writerMutex.Unlock()

	if writer == nil {

		writer = getInfluxClient2().WriteAPI("", InfluxGameDB+"/"+InfluxRetentionPolicyAllTime.String())

		go func() {
			for err := range writer.Errors() {
				log.Err("writing to influx", zap.Error(err))
			}
		}()
	}

	return writer
}
