package helpers

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gamedb/gamedb/pkg/log"
	"go.uber.org/zap"
)

func KeepAlive(callbacks ...func()) {

	var signals = []os.Signal{
		syscall.SIGTERM,
		syscall.SIGHUP,
		syscall.SIGQUIT,
		syscall.SIGINT,
		syscall.SIGKILL,
		os.Interrupt,
	}

	signalsChan := make(chan os.Signal, len(signals))
	signal.Notify(signalsChan, signals...)

	s := <-signalsChan // Blocks

	// Run callbacks
	var wg sync.WaitGroup
	for _, callback := range callbacks {
		wg.Add(1)
		go func(callback func()) {
			callback()
			wg.Done()
		}(callback)
	}
	wg.Wait()

	log.Info("Shutting down", zap.String("signal", s.String()))
	log.Flush()
}
