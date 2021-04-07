package charts

import (
	"bytes"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/Jleagle/influxql"
	"github.com/Jleagle/steam-go/steamapi"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/influx/schemas"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
)

func GetGroupChart(commandID string, groupID string, title string) (path string) {

	builder := influxql.NewBuilder()
	builder.AddSelect(`MAX("members_count")`, "max_members_count")
	builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementGroups.String())
	builder.AddWhere("time", ">", "NOW()-365d")
	builder.AddWhere("group_id", "=", groupID)
	builder.AddGroupByTime("1d")
	builder.SetFillNone()

	path, err := getInfluxChart(commandID, builder, groupID, title)
	if err != nil {
		log.Err(err.Error())
	}
	return path
}

func GetAppPlayersChart(commandID string, appID int, groupBy string, time string, title string) (path string) {

	builder := influxql.NewBuilder()
	if appID == 0 {
		builder.AddSelect("MAX(player_online)", "max_player_online")
	} else {
		builder.AddSelect("MAX(player_count)", "max_player_count")
	}
	builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementApps.String())
	builder.AddWhere("time", ">", "NOW()-"+time)
	builder.AddWhere("app_id", "=", appID)
	builder.AddGroupByTime(groupBy)
	builder.SetFillNone()

	path, err := getInfluxChart(commandID, builder, strconv.Itoa(appID), title)
	if err != nil {
		log.Err(err.Error())
	}
	return path
}

func GetPlayerChart(commandID string, playerID int64, field schemas.PlayerField, title string) (path string) {

	builder := influxql.NewBuilder()
	builder.AddSelect("MAX("+string(field)+")", "max_"+string(field))
	builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementPlayers.String())
	builder.AddWhere("time", ">", "NOW()-365d")
	builder.AddWhere("player_id", "=", playerID)
	builder.AddGroupByTime("1d")
	builder.SetFillNone()

	path, err := getInfluxChart(commandID, builder, strconv.FormatInt(playerID, 10), title)
	if err != nil {
		log.Err(err.Error())
	}
	return path
}

func getInfluxChart(commandID string, builder *influxql.Builder, id string, title string) (path string, err error) {

	resp, err := influx.InfluxQuery(builder)
	if err != nil {
		return "", err
	}

	if !(len(resp.Results) > 0 && len(resp.Results[0].Series) > 0) {
		return "", nil
	}

	x, y := influx.InfluxResponseToImageChartData(resp.Results[0].Series[0])

	return getChart(x, y, id, title, commandID)
}

func GetPriceChart(code steamapi.ProductCC, commandID string, id int, title string) (path string) {

	prices, err := mongo.GetPricesForProduct(id, helpers.ProductTypeApp, code)
	if err != nil {
		log.Err(err.Error())
		return ""
	}

	var x []time.Time
	var y []float64

	for k, price := range prices {

		value := float64(price.PriceAfter) / 100

		x = append(x, price.CreatedAt)
		y = append(y, value)

		// Create stepped chart
		if len(prices) > k+1 {
			x = append(x, prices[k+1].CreatedAt.Add(time.Second*-1))
			y = append(y, value)
		}
	}

	if len(y) > 0 {
		x = append(x, time.Now())
		y = append(y, y[len(y)-1])
	}

	path, err = getChart(x, y, strconv.Itoa(id), title, commandID)
	if err != nil {
		log.Err(err.Error())
		return ""
	}

	return path
}

func getChart(x []time.Time, y []float64, id string, title string, commandID string) (path string, err error) {

	if len(x) < 1 || len(y) < 1 {
		return "", nil
	}

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
			Name:     title,
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
				Min: 0,
				Max: max + 1,
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
		return "", err
	}

	var b = buffer.Bytes()
	var file = commandID + "-" + id + ".png"

	// Save chart to file
	if config.C.ChatBotAttachments == "" {
		return "", errors.New("missing environment variables")
	}

	f, err := os.OpenFile(config.C.ChatBotAttachments+"/"+file, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return "", err
	}

	defer helpers.Close(f)

	_, err = f.Write(b)
	if err != nil {
		log.ErrS(err)
	}

	return config.C.GlobalSteamDomain + "/assets/img/chatbot/" + file + "?_=" + strconv.FormatInt(time.Now().Truncate(time.Minute*10).Unix(), 10), err
}
