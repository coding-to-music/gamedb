package tasks

import (
	"time"

	influxHelper "github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
	influx "github.com/influxdata/influxdb1-client"
)

type GlobalSteamStats struct {
	BaseTask
}

func (c GlobalSteamStats) ID() string {
	return "gamedb-stats"
}

func (c GlobalSteamStats) Name() string {
	return "Game DB Stats"
}

func (c GlobalSteamStats) Group() TaskGroup {
	return ""
}

func (c GlobalSteamStats) Cron() TaskTime {
	return CronTimeGameDBStats
}

func (c GlobalSteamStats) work() (err error) {

	apps, err := mongo.CountDocuments(mongo.CollectionApps, nil, 0)
	if err != nil {
		return err
	}

	changes, err := mongo.CountDocuments(mongo.CollectionChanges, nil, 0)
	if err != nil {
		return err
	}

	groups, err := mongo.CountDocuments(mongo.CollectionGroups, nil, 0)
	if err != nil {
		return err
	}

	packages, err := mongo.CountDocuments(mongo.CollectionPackages, nil, 0)
	if err != nil {
		return err
	}

	webhooks, err := mongo.CountDocuments(mongo.CollectionWebhooks, nil, 0)
	if err != nil {
		return err
	}

	players, err := mongo.CountDocuments(mongo.CollectionPlayers, nil, 0)
	if err != nil {
		return err
	}

	users, err := mysql.CountUsers()
	if err != nil {
		return err
	}

	// Save to influx
	var point = influx.Point{
		Measurement: string(influxHelper.InfluxMeasurementGameDBStats),
		Fields: map[string]interface{}{
			"apps":     apps,
			"changes":  changes,
			"groups":   groups,
			"packages": packages,
			"webhooks": webhooks,
			"players":  players,
			"users":    users,
		},
		Time:      time.Now(),
		Precision: "h",
	}

	_, err = influxHelper.InfluxWrite(influxHelper.InfluxRetentionPolicyAllTime, point)
	return err
}
