package helpers

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func KeepAlive() {

	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGTERM, os.Interrupt)

	wg := &sync.WaitGroup{} // Must be pointer
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		for range signals {
			wg.Done()
		}
	}(wg)

	wg.Wait()
}
