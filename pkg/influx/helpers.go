package influx

import (
	"encoding/json"
	"math"
	"sort"
	"time"

	"github.com/Jleagle/influxql"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	influx "github.com/influxdata/influxdb1-client"
	"github.com/influxdata/influxdb1-client/models"
	influxModels "github.com/influxdata/influxdb1-client/models"
	"gonum.org/v1/gonum/stat"
)

const (
	InfluxGameDB     = "GameDB"
	InfluxTelegrafDB = "Telegraf-Web"

	InfluxRetentionPolicyAllTime InfluxRetentionPolicy = "alltime"
	InfluxRetentionPolicy14Day   InfluxRetentionPolicy = "14d"
	// InfluxRetentionPolicy7Day    InfluxRetentionPolicy = "7d"

	InfluxMeasurementAPICalls    InfluxMeasurement = "api_calls"
	InfluxMeasurementApps        InfluxMeasurement = "apps"
	InfluxMeasurementChanges     InfluxMeasurement = "changes"
	InfluxMeasurementChatBot     InfluxMeasurement = "chat_bot"
	InfluxMeasurementGroups      InfluxMeasurement = "groups"
	InfluxMeasurementPlayers     InfluxMeasurement = "players"
	InfluxMeasurementRabbitQueue InfluxMeasurement = "rabbitmq_queue"
	InfluxMeasurementSignups     InfluxMeasurement = "signups"
	// InfluxMeasurementPackages    InfluxMeasurement = "packages"
	// InfluxMeasurementStats       InfluxMeasurement = "stats"
	// InfluxMeasurementTags        InfluxMeasurement = "tags"
)

type InfluxRetentionPolicy string

func (irp InfluxRetentionPolicy) String() string {
	return string(irp)
}

type InfluxMeasurement string

func (im InfluxMeasurement) String() string {
	return string(im)
}

func InfluxWrite(retention InfluxRetentionPolicy, point influx.Point) (resp *influx.Response, err error) {
	return InfluxWriteMany(retention, influx.BatchPoints{
		Points: []influx.Point{point},
	})
}

func InfluxWriteMany(retention InfluxRetentionPolicy, batch influx.BatchPoints) (resp *influx.Response, err error) {

	if len(batch.Points) == 0 {
		return nil, nil
	}

	batch.Database = InfluxGameDB
	batch.RetentionPolicy = string(retention)
	batch.Precision = batch.Points[0].Precision // Must be in batch and point

	if batch.Time.IsZero() || batch.Time.Unix() == 0 {
		batch.Time = time.Now()
	}

	client, err := getInfluxClient()
	if err != nil {
		return nil, err
	}

	return client.Write(batch)
}

func InfluxQuery(builder *influxql.Builder) (resp *influx.Response, err error) {

	client, err := getInfluxClient()
	if err != nil {
		return resp, err
	}

	resp, err = client.Query(influx.Query{
		Command:         builder.String(),
		Database:        InfluxGameDB,
		RetentionPolicy: string(InfluxRetentionPolicyAllTime),
	})

	return resp, err
}

type (
	HighChartsJSON      map[string][][]interface{}
	HighChartsJSONMulti struct {
		Key   string         `json:"key"`
		Value HighChartsJSON `json:"value"`
	}
)

func InfluxResponseToHighCharts(series influxModels.Row, trimLeft bool) HighChartsJSON {

	resp := HighChartsJSON{}

	for k, v := range series.Columns {
		if k > 0 {

			var hasValue bool
			for _, vv := range series.Values {

				// Check if any of the series' have a value above zero
				if !hasValue && trimLeft {
					for k, vvv := range vv {
						if k > 0 {
							if val, ok := vvv.(json.Number); ok {
								i, err := val.Float64()
								if err == nil && i != 0 {
									hasValue = true
									break
								}
							}
						}
					}
				}

				if trimLeft && !hasValue {
					continue
				}

				t, err := time.Parse(time.RFC3339, vv[0].(string))
				if err != nil {
					log.ErrS(err)
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

func InfluxResponseToImageChartData(series influxModels.Row) (x []time.Time, y []float64) {

	var gotData bool // This is used to trim leading zeros

	for k := range series.Columns {
		if k == 1 {
			for _, vv := range series.Values {

				t, err := time.Parse(time.RFC3339, vv[0].(string))
				if err != nil {
					log.ErrS(err)
					continue
				}

				val, ok := vv[k].(json.Number)
				if ok {
					i, err := val.Float64()
					if err == nil {

						if i > 0 {
							gotData = true
						}

						if gotData {
							x = append(x, t)
							y = append(y, i)
						}
					}
				}
			}
		}
	}

	return x, y
}

func GetFirstInfluxInt(builder *influxql.Builder) (i int64, err error) {

	resp, err := InfluxQuery(builder)
	if err != nil {
		return 0, err
	}

	if resp != nil &&
		len(resp.Results) > 0 &&
		len(resp.Results[0].Series) > 0 &&
		len(resp.Results[0].Series[0].Values) > 0 &&
		len(resp.Results[0].Series[0].Values[0]) > 1 {

		if val, ok := resp.Results[0].Series[0].Values[0][1].(json.Number); ok {
			return val.Int64()
		}
	}

	return i, err
}

func GetInfluxTrendFromResponse(builder *influxql.Builder, padding int) (trend float64, err error) {

	resp, err := InfluxQuery(builder)
	if err != nil {
		return 0, err
	}

	if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 && len(resp.Results[0].Series[0].Values) > 0 {

		return GetInfluxTrendFromSeries(resp.Results[0].Series[0], padding), nil
	}

	return trend, nil
}

func GetInfluxTrendFromSeries(series models.Row, padding int) (trend float64) {

	var xs []float64
	var ys []float64

	for _, v := range series.Values {

		val, err := v[1].(json.Number).Int64()
		if err != nil {
			log.ErrS(err)
			continue
		}

		ys = append(ys, float64(val))
	}

	avg := helpers.Max(ys...)

	for k := range series.Values {
		xs = append(xs, float64(k)*avg)
	}

	if len(ys) > 0 {

		// Padding
		if padding > 0 && len(xs) < padding {
			diff := padding - len(xs)
			for i := 1; i <= diff; i++ {
				xs = append(xs, float64(len(xs)))    // Append
				ys = append([]float64{ys[0]}, ys...) // Prepend
			}
		}

		_, slope := stat.LinearRegression(xs, ys, nil, false)
		if !math.IsNaN(slope) {
			trend = slope
		}
	}

	return trend
}
