package helpers

import (
	"os"
	"os/signal"
	"syscall"
)

func KeepAlive() {

	var signals = []os.Signal{
		syscall.SIGTERM,
		os.Interrupt,
	}

	signalsChan := make(chan os.Signal, len(signals))
	signal.Notify(signalsChan, signals...)

	// Block
	<-signalsChan
}
