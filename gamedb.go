package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
)

const TMP = "/tmp/gamedb/"

var processes []string

func main() {

	err := config.Init("", "", "")
	log.InitZap(log.LogNameDevenv)
	defer log.Flush()
	if err != nil {
		log.ErrS(err)
		return
	}

	if len(os.Args) <= 1 {
		log.InfoS("please specify a cmd")
		return
	}
	processes = os.Args[1:]

	stopAll()
	startAll()

	go watchFiles()

	// ctrl-c
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	<-signals

	stopAll()
}

func startAll() {

	log.InfoS("starting: " + strings.Join(processes, ", "))

	for _, process := range processes {
		go func(process string) {

			cmd := exec.Command("sh", "-c", `cd ./cmd/`+process+`/; go build -ldflags "-X main.version=$(git rev-parse --verify HEAD) -X main.commits=$(git rev-list --count master)" *.go; ./`+process+``)
			err := cmd.Start()
			if err != nil {
				log.ErrS(err)
				return
			}

			filename := TMP + process + ".pid"

			err = ioutil.WriteFile(filename, []byte(strconv.Itoa(cmd.Process.Pid)), 0644)
			if err != nil {
				log.ErrS(err)
				return
			}

		}(process)
	}
}

func stopAll() {

	log.InfoS("quitting: " + strings.Join(processes, ", "))

	for _, process := range processes {

		filename := TMP + process + ".pid"

		b, err := ioutil.ReadFile(filename)
		if err != nil {
			continue
		}

		exec.Command("sh", "-c", "kill", string(b))

		err = os.Remove(filename)
		if err != nil {
			log.ErrS(err)
			continue
		}
	}
}

func watchFiles() {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.ErrS(err)
		return
	}
	defer helpers.Close(watcher)

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if event.Op&fsnotify.Write == fsnotify.Write {

					if strings.HasPrefix(filepath.Ext(event.Name), ".go") {

						log.InfoS("Updating: ", event.Name)
						stopAll()
						startAll()
					}
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}

				log.ErrS(err)
			}
		}
	}()

	dirs, err := getDirs()
	if err != nil {
		log.ErrS(err)
		return
	}

	log.InfoS("waching ", len(dirs), " dirs")

	for _, v := range dirs {
		err = watcher.Add(v)
		if err != nil {
			log.ErrS(err)
		}
	}
	<-done
}

func getDirs() (p []string, err error) {

	p = append(p, ".")

	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}
		if strings.HasPrefix(path, "node_modules") || strings.HasPrefix(path, ".") {
			return nil
		}

		p = append(p, path)

		return nil
	})

	return p, err
}
