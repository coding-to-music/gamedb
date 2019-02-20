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

	InfluxTableApps     InfluxTable = "apps"
	InfluxTablePackages InfluxTable = "packages"
	InfluxTableTags     InfluxTable = "tags"
	InfluxTablePlayers  InfluxTable = "players"
	InfluxTableStats    InfluxTable = "stats"
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
