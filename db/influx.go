package db

import (
	"net/url"
	"sync"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/gamedb/website/config"
	"github.com/gamedb/website/log"
	influx "github.com/influxdata/influxdb1-client"
	"github.com/influxdata/influxdb1-client/models"
)

type InfluxRetentionPolicy string
type InfluxMeasurement string

const (
	InfluxDB = "GameDB"

	InfluxRetentionPolicyAllTime InfluxRetentionPolicy = "alltime"
	InfluxRetentionPolicy7Day    InfluxRetentionPolicy = "7d"

	InfluxMeasurementApps     InfluxMeasurement = "apps"
	InfluxMeasurementPackages InfluxMeasurement = "packages"
	InfluxMeasurementTags     InfluxMeasurement = "tags"
	InfluxMeasurementPlayers  InfluxMeasurement = "players"
	InfluxMeasurementStats    InfluxMeasurement = "stats"
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

func InfluxWrite(retention InfluxRetentionPolicy, point influx.Point) (resp *influx.Response, err error) {

	return InfluxWriteMany(retention, influx.BatchPoints{
		Points: []influx.Point{point},
	})
}

func InfluxWriteMany(retention InfluxRetentionPolicy, batch influx.BatchPoints) (resp *influx.Response, err error) {

	batch.Database = InfluxDB
	batch.RetentionPolicy = string(retention)
	batch.Precision = "m" // Must be in batch and point

	if batch.Time.IsZero() {
		batch.Time = time.Now()
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

func InfluxQuery(query string) (resp *influx.Response, err error) {

	client, err := GetInfluxClient()
	if err != nil {
		return &influx.Response{}, err
	}

	resp, err = client.Query(influx.Query{
		Command:         query,
		Database:        InfluxDB,
		RetentionPolicy: string(InfluxRetentionPolicyAllTime),
	})

	if len(resp.Results) == 0 {
		log.Warning(log.ServiceLocal, "No results returned: "+query)
	}

	if len(resp.Results) > 0 && len(resp.Results[0].Series) == 0 {
		log.Warning(log.ServiceLocal, "No series returned: "+query)
	}

	return resp, err
}

type HighChartsJson map[string][][]interface{}

func InfluxResponseToHighCharts(series models.Row) HighChartsJson {

	json := HighChartsJson{}

	for k, v := range series.Columns {
		if k > 0 {
			for _, vv := range series.Values {
				t, err := time.Parse(time.RFC3339, vv[0].(string))
				if err != nil {
					log.Err(err)
				}

				json[v] = append(json[v], []interface{}{t.Unix() * 1000, vv[k]})
			}
		}
	}

	return json
}
