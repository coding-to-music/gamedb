package main

import (
	"github.com/gamedb/website/pkg"
)

func main() {

	pkg.Info("Starting consumers")

	if config.Config.EnableConsumers.GetBool() {
		go func() {
			pkg.Info("Starting consumers")
			RunConsumers()
		}()
	}

	pkg.KeepAlive()
}
