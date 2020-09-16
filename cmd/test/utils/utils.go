package utils

import (
	"os"

	"github.com/gamedb/gamedb/pkg/log"
	"go.uber.org/zap"
)

type util interface {
	name() string
	run()
}

var utils = []util{
	queuePackages{},
	saveFromPics{},
	syncStates{},
	testUtil{},
}

func init() {
	m := make(map[string]bool, len(utils))
	for _, v := range utils {
		if _, ok := m[v.name()]; ok {
			log.Err("Duplicate util name", zap.String("util", v.name()))
			os.Exit(1)
		}
		m[v.name()] = true
	}
}

func RunUtil(util string) {
	for _, v := range utils {
		if util == v.name() {
			v.run()
			os.Exit(0)
		}
	}
}
