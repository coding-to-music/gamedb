package tasks

import (
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	influx "github.com/influxdata/influxdb1-client"
)

type SteamClientPlayers struct {
}

func (c SteamClientPlayers) ID() string {
	return "update-steam-client-players"
}

func (c SteamClientPlayers) Name() string {
	return "Update Steam client players"
}

func (c SteamClientPlayers) Cron() string {
	return "*/10 *"
}

func (c SteamClientPlayers) work() {

	resp, err := http.Get("https://www.valvesoftware.com/en/about/stats")
	if err != nil {
		log.Err(err)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Err(err)
		return
	}

	if strings.TrimSpace(string(b)) == "[]" {
		log.Info("www.valvesoftware.com/en/about/stats returned empty array")
		return
	}

	sp := steamPlayersStruct{}
	err = helpers.Unmarshal(b, &sp)
	if err != nil {
		log.Warning("www.valvesoftware.com/en/about/stats down: " + string(b))
		return
	}

	fields := map[string]interface{}{
		"player_online": sp.int(sp.Online),
		"player_count":  sp.int(sp.InGame),
	}

	_, err = helpers.InfluxWrite(helpers.InfluxRetentionPolicyAllTime, influx.Point{
		Measurement: string(helpers.InfluxMeasurementApps),
		Tags: map[string]string{
			"app_id": "0",
		},
		Fields:    fields,
		Time:      time.Now(),
		Precision: "m",
	})

	log.Warning(err)
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
