package helpers

import (
	"encoding/json"
	"net/url"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	influx "github.com/influxdata/influxdb1-client"
	"github.com/influxdata/influxdb1-client/models"
)

type InfluxRetentionPolicy string

func (irp InfluxRetentionPolicy) String() string {
	return string(irp)
}

type InfluxMeasurement string

func (im InfluxMeasurement) String() string {
	return string(im)
}

const (
	InfluxGameDB     = "GameDB"
	InfluxTelegrafDB = "Telegraf"

	InfluxRetentionPolicyAllTime InfluxRetentionPolicy = "alltime"
	InfluxRetentionPolicy7Day    InfluxRetentionPolicy = "7d"
	InfluxRetentionPolicy14Day   InfluxRetentionPolicy = "14d"

	InfluxMeasurementApps        InfluxMeasurement = "apps"
	InfluxMeasurementPackages    InfluxMeasurement = "packages"
	InfluxMeasurementTags        InfluxMeasurement = "tags"
	InfluxMeasurementPlayers     InfluxMeasurement = "players"
	InfluxMeasurementStats       InfluxMeasurement = "stats"
	InfluxMeasurementRabbitQueue InfluxMeasurement = "rabbitmq_queue"
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
		host, err = url.Parse(config.Config.InfluxURL.Get())
		if err != nil {
			return
		}

		conf := influx.Config{
			URL:      *host,
			Username: config.Config.InfluxUsername.Get(),
			Password: config.Config.InfluxPassword.Get(),
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

	batch.Database = InfluxGameDB
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
		Database:        InfluxGameDB,
		RetentionPolicy: string(InfluxRetentionPolicyAllTime),
	})

	return resp, err
}

type HighChartsJson map[string][][]interface{}

func InfluxResponseToHighCharts(series models.Row) HighChartsJson {

	resp := HighChartsJson{}

	for k, v := range series.Columns {
		if k > 0 {
			for _, vv := range series.Values {
				t, err := time.Parse(time.RFC3339, vv[0].(string))
				if err != nil {
					log.Err(err)
					continue
				}

				resp[v] = append(resp[v], []interface{}{t.Unix() * 1000, vv[k]})
			}
		}
	}

	for k := range resp {

		sort.Slice(resp[k], func(i, j int) bool {
			return resp[k][i][0].(int64) < resp[k][j][0].(int64)
		})

	}

	return resp
}

func GetFirstInfluxInt(resp *influx.Response) int {

	if resp != nil &&
		len(resp.Results) > 0 &&
		len(resp.Results[0].Series) > 0 &&
		len(resp.Results[0].Series[0].Values) > 0 &&
		len(resp.Results[0].Series[0].Values[0]) > 1 {

		switch v := resp.Results[0].Series[0].Values[0][1].(type) {
		case int:
			return v
		case json.Number:
			i, err := v.Int64()
			log.Err(err)
			return int(i)
		default:
			log.Warning("Unknown type from Influx DB: " + reflect.TypeOf(v).String())
		}
	}

	return 0
}
