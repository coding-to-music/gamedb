package chatbot

import (
	"bytes"
	"time"

	"github.com/Jleagle/influxql"
	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/elastic"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/influx"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
)

type CommandGroup struct {
}

func (CommandGroup) Regex() string {
	return `^[.|!](group|clan) (.*)`
}

func (CommandGroup) DisableCache() bool {
	return false
}

func (CommandGroup) Example() string {
	return ".group GroupName"
}

func (CommandGroup) Description() string {
	return "Get info on a group"
}

func (CommandGroup) Type() CommandType {
	return TypeGroup
}

func (c CommandGroup) Output(msg *discordgo.MessageCreate) (message discordgo.MessageSend, err error) {

	matches := RegexCache[c.Regex()].FindStringSubmatch(msg.Message.Content)

	groups, _, _, err := elastic.SearchGroups(0, 1, nil, matches[2], "")
	if err != nil {
		return message, err
	} else if len(groups) == 0 {
		message.Content = "Group **" + matches[2] + "** not found"
		return message, nil
	}

	group := groups[0]

	var abbr = group.GetAbbr()
	if abbr == "" {
		abbr = "-"
	}
	var headline = group.GetHeadline()
	if headline == "" {
		headline = "-"
	}

	var image *discordgo.MessageEmbedImage
	url, width, height, err := getGroupChart(group.ID)
	if err != nil {
		log.Err(err)
	} else if url != "" {
		image = &discordgo.MessageEmbedImage{
			URL:    url,
			Width:  width,
			Height: height,
		}
	}

	message.Content = "<@" + msg.Author.ID + ">"
	message.Embed = &discordgo.MessageEmbed{
		Image: image,
		Title: group.GetName(),
		URL:   "https://gamedb.online" + group.GetPath(),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: group.GetIcon(),
		},
		Footer: getFooter(),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Headline",
				Value: headline,
			},
			{
				Name:  "Short Name",
				Value: abbr,
			},
			{
				Name:  "Members",
				Value: humanize.Comma(int64(group.Members)),
			},
		},
	}

	return message, nil
}

func getGroupChart(id string) (url string, width int, height int, err error) {

	builder := influxql.NewBuilder()
	builder.AddSelect(`max("members_count")`, "max_members_count")
	builder.SetFrom(influx.InfluxGameDB, influx.InfluxRetentionPolicyAllTime.String(), influx.InfluxMeasurementGroups.String())
	builder.AddWhere("group_id", "=", id)
	builder.AddGroupByTime("1d")
	builder.SetFillNone()

	resp, err := influx.InfluxQuery(builder.String())
	if err != nil {
		return "", 0, 0, err
	}

	if !(len(resp.Results) > 0 && len(resp.Results[0].Series) > 0) {
		return "", 0, 0, err
	}

	x, y := influx.InfluxResponseToImageChartData(resp.Results[0].Series[0])

	min := helpers.Min(y...) - 1
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
				return humanize.Commaf(v.(float64))
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
	var file = "group-" + id + ".png"
	var b = buffer.Bytes()

	if config.IsLocal() {
		err = saveChartToFile(b, file)
		if err != nil {
			return "", 0, 0, err
		}
	}

	u, err := saveChartToGoogle(b, file)
	return u, graph.GetWidth(), graph.GetHeight(), err
}
