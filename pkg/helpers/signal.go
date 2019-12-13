package helpers

import (
	"os"
	"os/signal"
	"syscall"
)

func KeepAlive() {

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGTERM, os.Interrupt)

	select {
	case <-signals:
	}
}
