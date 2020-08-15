package charts

import (
	"bytes"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/Jleagle/influxql"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
	"go.uber.org/zap"
)

func GetGroupChart(group elasticsearch.Group) (reader io.Reader, err error) {

	builder := influxql.NewBuilder()
	builder.AddSelect(`max("members_count")`, "max_members_count")
	builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementGroups.String())
	builder.AddWhere("group_id", "=", group.ID)
	builder.AddGroupByTime("1d")
	builder.SetFillNone()

	return getChart(builder, group.ID, "Members", config.Config.GameDBDomain.Get()+group.GetPath())
}

func GetAppChart(app mongo.App) (reader io.Reader, err error) {

	builder := influxql.NewBuilder()
	builder.AddSelect("max(player_count)", "max_player_count")
	builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementApps.String())
	builder.AddWhere("app_id", "=", app.ID)
	builder.AddGroupByTime("1d")
	builder.SetFillNumber(0)

	return getChart(builder, strconv.Itoa(app.ID), "In Game", config.Config.GameDBDomain.Get()+app.GetPath())
}

func getChart(builder *influxql.Builder, id string, title string, description string) (reader io.Reader, err error) {

	resp, err := influx.InfluxQuery(builder.String())
	if err != nil {
		return nil, err
	}

	if !(len(resp.Results) > 0 && len(resp.Results[0].Series) > 0) {
		return nil, nil
	}

	x, y := influx.InfluxResponseToImageChartData(resp.Results[0].Series[0])

	min := helpers.Max(helpers.Min(y...)-1, 0)
	max := helpers.Max(y...) + 1

	if len(x) == 1 {
		x = append(x, x[0].Add(-time.Hour))
		y = append(y, y[0])
	}

	var (
		colourDark  = drawing.ColorFromHex("1b2738")
		colourLight = drawing.ColorFromHex("e9ecef")
		colourGreen = drawing.ColorFromHex("28a745")
	)

	graph := chart.Chart{
		Title: title,
		TitleStyle: chart.Style{
			Show:      true,
			FontColor: colourLight,
		},
		Background: chart.Style{
			FillColor: colourDark,
		},
		Canvas: chart.Style{
			FillColor: colourDark,
		},
		XAxis: chart.XAxis{
			Style: chart.Style{
				Show:        true,
				FontColor:   colourLight,
				StrokeColor: colourLight,
			},
			ValueFormatter: chart.TimeDateValueFormatter,
		},
		YAxis: chart.YAxis{
			Name:     "Members",
			AxisType: chart.YAxisSecondary,
			Style: chart.Style{
				Show:        true,
				FontColor:   colourLight,
				StrokeColor: colourLight,
			},
			GridMajorStyle: chart.Style{
				Show: true,
			},
			Range: &chart.ContinuousRange{
				Min: min,
				Max: max,
			},
			ValueFormatter: func(v interface{}) string {
				if (max - min) > 10 {
					return humanize.Comma(int64(v.(float64)))
				}
				return humanize.Commaf(helpers.RoundFloatTo2DP(v.(float64)))
			},
		},
		Series: []chart.Series{
			chart.TimeSeries{
				YAxis: chart.YAxisPrimary,
				Style: chart.Style{
					Show:        true,
					StrokeColor: colourGreen,
					StrokeWidth: 2,
					FillColor:   colourDark,
				},
				XValues: x,
				YValues: y,
			},
		},
	}

	buffer := bytes.NewBuffer([]byte{})
	err = graph.Render(chart.PNG, buffer)
	if err != nil {
		return nil, err
	}

	var b = buffer.Bytes()

	// Save chart to file
	if config.IsLocal() {

		f, err := os.OpenFile("app-"+id+".png", os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			return nil, err
		}

		defer func() {
			err := f.Close()
			zap.S().Error(err)
		}()

		_, err = f.Write(b)
		if err != nil {
			return nil, err
		}
	}

	return bytes.NewReader(b), err
}
