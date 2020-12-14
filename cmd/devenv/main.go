package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
)

const TMP = "/tmp/gamedb/"

type processesConfig struct {
	Processes []struct {
		Path    string
		Enabled bool
	}
	Mongo    bool
	Elastic  bool
	Memcache bool
	MySQL    bool
	Rabbit   bool
}

func main() {

	err := config.Init("")
	log.InitZap(log.LogNameDevenv)
	defer log.Flush()
	if err != nil {
		log.ErrS(err)
		return
	}

	// Get config
	b, err := ioutil.ReadFile("./cmd/devenv/config.json")
	if err != nil {
		log.ErrS(err)
		return
	}

	var cfg processesConfig
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		log.ErrS(err)
		return
	}

	stopAll(cfg)
	startAll(cfg)

	go watchFiles(cfg)

	helpers.KeepAlive()

	stopAll(cfg)
}

func startAll(cfg processesConfig) {

	log.InfoS("starting")

	for _, process := range cfg.Processes {
		if process.Enabled {
			go func(process string) {

				cmd := exec.Command("sh", "-c", `go build -o `+process+` -ldflags "-X main.version=$(git rev-parse --verify HEAD) -X main.commits=$(git rev-list --count master)" *.go`)
				cmd.Dir = "./cmd/" + process
				err := cmd.Run() // Blocks
				if err != nil {
					log.ErrS(err)
					return
				}

				cmd = exec.Command("sh", "-c", "./"+process)
				cmd.Dir = "./cmd/" + process
				err = cmd.Start()
				if err != nil {
					log.ErrS(err)
					return
				}

				err = ioutil.WriteFile(TMP+process+".pid", []byte(strconv.Itoa(cmd.Process.Pid)), 0644)
				if err != nil {
					log.ErrS(err)
				}

			}(process.Path)
		}
	}
}

func stopAll(cfg processesConfig) {

	log.InfoS("quitting")

	for _, process := range cfg.Processes {

		if process.Enabled {

			filename := TMP + process.Path + ".pid"

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
}

var lastUpdated time.Time

func watchFiles(cfg processesConfig) {

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

						if time.Now().Sub(lastUpdated) > (time.Second / 10) {
							lastUpdated = time.Now()
							log.InfoS("Updating: ", event.Name)
							stopAll(cfg)
							startAll(cfg)
						}
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
