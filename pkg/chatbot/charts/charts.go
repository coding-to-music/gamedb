package charts

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"cloud.google.com/go/storage"
	"github.com/Jleagle/influxql"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/elasticsearch"
	"github.com/gamedb/gamedb/pkg/googlestorage"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
)

func GetGroupChart(group elasticsearch.Group) (url string, width int, height int, err error) {

	builder := influxql.NewBuilder()
	builder.AddSelect(`max("members_count")`, "max_members_count")
	builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementGroups.String())
	builder.AddWhere("group_id", "=", group.ID)
	builder.AddGroupByTime("1d")
	builder.SetFillNone()

	return getChart(builder, group.ID, "Members", "https://gamedb.online"+group.GetPath())
}

func GetAppChart(app mongo.App) (url string, width int, height int, err error) {

	builder := influxql.NewBuilder()
	builder.AddSelect("max(player_count)", "max_player_count")
	builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementApps.String())
	builder.AddWhere("app_id", "=", app.ID)
	builder.AddGroupByTime("1d")
	builder.SetFillNumber(0)

	return getChart(builder, strconv.Itoa(app.ID), "In Game", "https://gamedb.online"+app.GetPath())
}

func getChart(builder *influxql.Builder, id string, title string, description string) (url string, width int, height int, err error) {

	resp, err := influx.InfluxQuery(builder.String())
	if err != nil {
		return "", 0, 0, err
	}

	if !(len(resp.Results) > 0 && len(resp.Results[0].Series) > 0) {
		return "", 0, 0, err
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
		return "", 0, 0, err
	}

	//
	var file = "app-" + id + ".png"
	var b = buffer.Bytes()

	if config.IsLocal() {
		err = saveChartToFile(b, file)
		if err != nil {
			return "", 0, 0, err
		}
	}

	// u, err := saveChartToGoogle(b, file)
	u, err := saveChartToImgur(b, title, description)
	return u, graph.GetWidth(), graph.GetHeight(), err
}

func saveChartToFile(b []byte, filename string) error {

	f, err := os.OpenFile(filename, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	defer func() {
		err := f.Close()
		log.Err(err)
	}()

	_, err = f.Write(b)
	return err
}

//noinspection GoUnusedFunction
func saveChartToGoogle(b []byte, filename string) (string, error) {

	client, ctx, err := googlestorage.GetStorageClient()
	if err != nil {
		return "", err
	}

	w := client.Bucket(googlestorage.BucketChatBot).Object(filename).NewWriter(ctx)

	_, err = io.Copy(w, bytes.NewBuffer(b))
	if err != nil {
		return "", err
	}
	err = w.Close()
	if err != nil {
		return "", err
	}

	opts, err := googlestorage.GetSignedURLOptions()
	if err != nil {
		return "", err
	}

	return storage.SignedURL(googlestorage.BucketChatBot, filename, opts)
}

// todo, 1250/month limit
func saveChartToImgur(b []byte, title, description string) (string, error) {

	headers := http.Header{}
	headers.Add("Authorization", "Client-ID "+config.Config.ImgurClientID.Get())
	headers.Add("Content-Type", "application/x-www-form-urlencoded")

	data := url.Values{}
	data.Add("image", string(b))
	data.Add("type", "file")
	data.Add("title", title)
	data.Add("description", description)

	b, code, err := helpers.Post("https://api.imgur.com/3/image", data, headers)
	if code != 200 || err != nil {
		return "", err
	}

	resp := ImgurResponse{}
	err = json.Unmarshal(b, &resp)
	return resp.Data.Link, err
}

type ImgurResponse struct {
	Data struct {
		Link string `json:"link"`
	} `json:"data"`
	Success bool `json:"success"`
	Status  int  `json:"status"`
}
