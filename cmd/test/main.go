package main

import (
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/sql"
)

func main() {

	config.SetVersion("test")
	log.Initialise([]log.LogName{log.LogNameTest}, "test")

	sApp, err := sql.GetApp(787860, nil)
	log.Err(err)

	err = sApp.SaveToMongo()
	log.Err(err)
}
