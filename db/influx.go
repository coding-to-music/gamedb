package db

import (
	"net/url"

	"github.com/gamedb/website/config"
	influx "github.com/influxdata/influxdb1-client"
)

type InfluxTable string

const (
	InfluxDB = "GameDB"

	InfluxTableApps     InfluxTable = "apps"
	InfluxTablePackages InfluxTable = "packages"
	InfluxTableTags     InfluxTable = "tags"
	InfluxTablePlayers  InfluxTable = "players"
	InfluxTableStats    InfluxTable = "stats"
)

func GetInfluxClient() (client *influx.Client, err error) {

	host, err := url.Parse(config.Config.InfluxHost)
	if err != nil {
		return
	}

	conf := influx.Config{
		URL:      *host,
		Username: "token",
		Password: config.Config.InfluxPassword,
	}

	return influx.NewClient(conf)
}
