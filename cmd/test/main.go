package main

import (
	"os"

	"github.com/gamedb/gamedb/cmd/test/utils"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/queue"
)

var version string
var commits string

func main() {

	err := config.Init(version, commits, helpers.GetIP())
	log.InitZap(log.LogNameTest)
	defer log.Flush()
	if err != nil {
		log.ErrS(err)
		return
	}

	if len(os.Args) > 1 {
		utils.RunUtil(os.Args[1])
	}

	queue.Init(queue.AllProducerDefinitions)

	//

	helpers.KeepAlive()
}
