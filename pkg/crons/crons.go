package crons

import (
	"strconv"

	"github.com/gamedb/gamedb/pkg/log"
)

// Logging
func cronLogErr(interfaces ...interface{}) {
	log.Err(append(interfaces, log.LogNameCron, log.LogNameGameDB)...)
}

func cronLogInfo(interfaces ...interface{}) {
	log.Info(append(interfaces, log.LogNameCron, log.LogNameGameDB)...)
}

func statsLogger(tableName string, count int, total int, rowName string) {

	cronLogInfo("Updating " + tableName + " - " + strconv.Itoa(count) + " / " + strconv.Itoa(total) + ": " + rowName)
}
