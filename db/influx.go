package db

import (
	"net/url"
	"sync"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/gamedb/website/config"
	"github.com/gamedb/website/log"
	influx "github.com/influxdata/influxdb1-client"
)

const (
	InfluxDB              = "GameDB"
	InfluxRetentionPolicy = "autogen"

	InfluxMeasurementApps     = "apps"
	InfluxMeasurementPackages = "packages"
	InfluxMeasurementTags     = "tags"
	InfluxMeasurementPlayers  = "players"
	InfluxMeasurementStats    = "stats"
)

var (
	influxClient *influx.Client
	influxLock   sync.Mutex
)

func GetInfluxClient() (client *influx.Client, err error) {

	influxLock.Lock()
	defer influxLock.Unlock()

	if influxClient == nil {

		var host *url.URL
		host, err = url.Parse(config.Config.InfluxURL)
		if err != nil {
			return
		}

		conf := influx.Config{
			URL:      *host,
			Username: config.Config.InfluxUsername,
			Password: config.Config.InfluxPassword,
		}

		influxClient, err = influx.NewClient(conf)
	}

	return influxClient, err
}

func InfluxWrite(point influx.Point) (resp *influx.Response, err error) {

	return InfluxWriteMany([]influx.Point{point})
}

func InfluxWriteMany(points []influx.Point) (resp *influx.Response, err error) {

	batch := influx.BatchPoints{
		Points:          points,
		Database:        InfluxDB,
		RetentionPolicy: InfluxRetentionPolicy,
	}

	client, err := GetInfluxClient()
	if err != nil {
		return &influx.Response{}, err
	}

	policy := backoff.NewExponentialBackOff()
	policy.InitialInterval = 1

	operation := func() (err error) {

		resp, err = client.Write(batch)
		return err
	}

	err = backoff.RetryNotify(operation, backoff.WithMaxRetries(policy, 5), func(err error, t time.Duration) { log.Info(err) })
	return resp, err
}

func InfluxResponseToHighCharts(resp *influx.Response) HighChartsJson {

	json := HighChartsJson{}

	if resp != nil {
		if len(resp.Results) > 0 {
			if len(resp.Results[0].Series) > 0 {

				for k, v := range resp.Results[0].Series[0].Columns {
					if k > 0 {

						var data []interface{}

						for _, vv := range resp.Results[0].Series[0].Values {
							data = append(data, []interface{}{vv[0], vv[k]})
						}

						json[v] = data
					}
				}
			}
		}
	}

	return json
}
