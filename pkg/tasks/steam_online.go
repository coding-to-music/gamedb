package tasks

import (
	"errors"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/helpers/influx"
	"github.com/gamedb/gamedb/pkg/log"
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

func (c SteamOnline) Cron() string {
	return CronTimeSteamClientPlayers
}

func (c SteamOnline) work() (err error) {

	resp, err := helpers.GetWithTimeout("https://www.valvesoftware.com/en/about/stats", 0)
	if err != nil {
		return err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if strings.TrimSpace(string(b)) == "[]" {
		return TaskError{
			Err:  errors.New("www.valvesoftware.com/en/about/stats returned empty array"),
			Okay: true,
		}
	}

	sp := steamPlayersStruct{}
	err = helpers.Unmarshal(b, &sp)
	if err != nil {
		return TaskError{
			Err:  errors.New("www.valvesoftware.com/en/about/stats down: " + string(b)),
			Okay: true,
		}
	}

	fields := map[string]interface{}{
		"player_online": sp.int(sp.Online),
		"player_count":  sp.int(sp.InGame),
	}

	_, err = influxHelper.InfluxWrite(influxHelper.InfluxRetentionPolicyAllTime, influx.Point{
		Measurement: string(influxHelper.InfluxMeasurementApps),
		Tags: map[string]string{
			"app_id": "0",
		},
		Fields:    fields,
		Time:      time.Now(),
		Precision: "m",
	})

	return err
}

type steamPlayersStruct struct {
	Online string `json:"users_online"`
	InGame string `json:"users_ingame"`
}

func (sp steamPlayersStruct) int(s string) int {
	s = strings.ReplaceAll(s, ",", "")
	i, err := strconv.Atoi(s)
	log.Warning(err)
	return i
}
