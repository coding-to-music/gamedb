package main

import (
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
)

func main() {

	log.Initialise([]log.LogName{log.LogNameTest})

	// Get API key
	err := sql.GetAPIKey("test", false)
	if err != nil {
		log.Critical(err)
		return
	}

	select {}
}
