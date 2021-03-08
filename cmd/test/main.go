package main

import (
	"os"

	"github.com/gamedb/gamedb/cmd/test/utils"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/consumers"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/mysql"
)

func main() {

	err := config.Init(helpers.GetIP())
	log.InitZap(log.LogNameTest)
	defer log.Flush()
	if err != nil {
		log.ErrS(err)
		return
	}

	if len(os.Args) > 1 {
		utils.RunUtil(os.Args[1])
	}

	consumers.Init(consumers.AllProducerDefinitions)

	//
	helpers.KeepAlive(
		mysql.Close,
		mongo.Close,
		memcache.Close,
	)
}
