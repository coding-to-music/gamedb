package main

import (
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
)

func main() {

	config.SetVersion("")
	log.Initialise([]log.LogName{log.LogNameTest})

	// Get API key
	err := sql.GetAPIKey("test", true)
	if err != nil {
		log.Critical(err)
		return
	}

	helpers.KeepAlive()
}
