package utils

import (
	"github.com/gamedb/gamedb/pkg/log"
)

type testUtil struct{}

func (testUtil) name() string {
	return "test"
}

func (testUtil) run() {
	log.Info("test")
}
