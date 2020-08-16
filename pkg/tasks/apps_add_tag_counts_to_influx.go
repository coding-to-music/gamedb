package tasks

import (
	"strconv"
	"time"

	"github.com/gamedb/gamedb/pkg/config"
	influxHelper "github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/mongo"
	influx "github.com/influxdata/influxdb1-client"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type AppsAddTagCountsToInflux struct {
	BaseTask
}

func (c AppsAddTagCountsToInflux) ID() string {
	return "add-tag-counts-to-influx"
}

func (c AppsAddTagCountsToInflux) Name() string {
	return "Add Tag Counts to Influx"
}

func (c AppsAddTagCountsToInflux) Group() TaskGroup {
	return TaskGroupApps
}

func (c AppsAddTagCountsToInflux) Cron() TaskTime {
	return CronTimeAddAppTagsToInflux
}

func (c AppsAddTagCountsToInflux) work() (err error) {

	var projection = bson.M{
		"_id":        1,
		"tag_counts": 1,
	}

	var filter = bson.D{
		{"tag_counts", bson.M{"$ne": bson.A{}}},
		{"updated_at", bson.M{"$gte": time.Now().Add(time.Hour * 24 * -1)}},
	}

	var i = 1
	return mongo.BatchApps(filter, projection, func(apps []mongo.App) {

		if config.IsLocal() {
			zap.S().Info(i)
			i++
		}

		var batch = influx.BatchPoints{
			Database:        influxHelper.InfluxGameDB,
			RetentionPolicy: influxHelper.InfluxRetentionPolicyAllTime.String(),
		}

		for _, app := range apps {

			if len(app.TagCounts) == 0 {
				continue
			}

			var fields = map[string]interface{}{}
			for _, v := range app.TagCounts {
				fields["tag_"+strconv.Itoa(v.ID)] = v.Count
			}

			batch.Points = append(batch.Points, influx.Point{
				Measurement: string(influxHelper.InfluxMeasurementApps),
				Tags: map[string]string{
					"app_id": strconv.Itoa(app.ID),
				},
				Fields:    fields,
				Time:      time.Now(),
				Precision: "h",
			})
		}

		_, err = influxHelper.InfluxWriteMany(influxHelper.InfluxRetentionPolicyAllTime, batch)
		if err != nil {
			zap.S().Error(err)
		}
	})
}
