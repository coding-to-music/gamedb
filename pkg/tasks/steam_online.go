package tasks

import (
	"errors"
	"strings"
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/influx"
	influx "github.com/influxdata/influxdb1-client"
)

type SteamOnline struct {
	BaseTask
}

func (c SteamOnline) ID() string {
	return "update-steam-client-players"
}

func (c SteamOnline) Name() string {
	return "Update Steam client players"
}

func (c SteamOnline) Group() TaskGroup {
	return ""
}

func (c SteamOnline) Cron() TaskTime {
	return CronTimeSteamClientPlayers
}

func (c SteamOnline) work() (err error) {

	body, _, err := helpers.Get("https://www.valvesoftware.com/en/about/stats", 0, nil)
	if err != nil {
		return err
	}

	if strings.TrimSpace(string(body)) == "[]" {
		return TaskError{
			Err:  errors.New("www.valvesoftware.com/en/about/stats returned empty array"),
			Okay: true,
		}
	}

	sp := steamPlayersStruct{}
	err = helpers.Unmarshal(body, &sp)
	if err != nil {
		return TaskError{
			Err:  errors.New("www.valvesoftware.com/en/about/stats down: " + string(body)),
			Okay: true,
		}
	}

	_, err = influxHelper.InfluxWrite(influxHelper.InfluxRetentionPolicyAllTime, influx.Point{
		Measurement: string(influxHelper.InfluxMeasurementApps),
		Tags: map[string]string{
			"app_id": "0",
		},
		Fields: map[string]interface{}{
			"player_online": helpers.StringToInt(sp.Online),
			"player_count":  helpers.StringToInt(sp.InGame),
		},
		Time:      time.Now(),
		Precision: "m",
	})

	return err
}

type steamPlayersStruct struct {
	Online string `json:"users_online"`
	InGame string `json:"users_ingame"`
}

// func fallback(){
//
// 	col := colly.NewCollector(
// 		steam.WithTimeout(0),
// 	)
//
// 	col.OnHTML("#stats_users_online", func(e *colly.HTMLElement) {
// 		online = toInt(e.Text)
// 	})
//
// 	col.OnHTML("#stats_users_ingame", func(e *colly.HTMLElement) {
// 		ingame = toInt(e.Text)
// 	})
//
// 	//
// 	col.OnError(func(r *colly.Response, err error) {
// 		log.ErrS(err)
// 	})
//
// 	err = col.Visit("https://www.valvesoftware.com/en/about")
// 	if err != nil {
// 		return err
// 	}
// }
