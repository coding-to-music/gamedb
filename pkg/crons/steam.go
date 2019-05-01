package crons

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

func SteamClientPlayers() {

	log.Info("Cron running: Steam users")

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

	sp := steamPlayersStruct{}
	err = helpers.Unmarshal(b, &sp)
	if err != nil {
		log.Err("www.valvesoftware.com/en/about/stats down")
		return
	}

	online := sp.int(sp.Online)
	inGame := sp.int(sp.InGame)

	fields := map[string]interface{}{
		"player_online": online,
	}

	// Sometimes ingames shows up as something close to online
	if inGame < online-1000000 {
		fields["player_count"] = inGame
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
