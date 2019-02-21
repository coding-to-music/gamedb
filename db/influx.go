package db

import (
	"net/url"

	"github.com/gamedb/website/config"
	influx "github.com/influxdata/influxdb1-client"
)

type InfluxTable string

const (
	InfluxDB              = "GameDB"
	InfluxRetentionPolicy = "autogen"

	InfluxMeasurementApps     = "apps"
	InfluxMeasurementPackages = "packages"
	InfluxMeasurementTags     = "tags"
	InfluxMeasurementPlayers  = "players"
	InfluxMeasurementStats    = "stats"
)

func GetInfluxClient() (client *influx.Client, err error) {

	host, err := url.Parse(config.Config.InfluxURL)
	if err != nil {
		return
	}

	conf := influx.Config{
		URL:      *host,
		Username: config.Config.InfluxUsername,
		Password: config.Config.InfluxPassword,
	}

	return influx.NewClient(conf)
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

	return client.Write(batch)
}

type HighChartsJson map[string][]interface{}

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
