package helpers

import (
	"os"
	"os/signal"
	"syscall"
)

func KeepAlive() {

	var signals = []os.Signal{
		syscall.SIGTERM,
		syscall.SIGHUP,
		syscall.SIGQUIT,
		syscall.SIGINT,
		syscall.SIGKILL,
	}

	signalsChan := make(chan os.Signal, len(signals))
	signal.Notify(signalsChan, signals...)

	// Block
	<-signalsChan
}
