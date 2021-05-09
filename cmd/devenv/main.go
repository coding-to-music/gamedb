package main

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"go.uber.org/zap"
)

const TMP = "/tmp/gamedb"

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
	b, err := os.ReadFile("./cmd/devenv/config.json")
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

	// Check Dependencies
	err = checkDependencies(cfg)
	if err != nil {
		log.ErrS(err)
		return
	}

	// Run
	stopAll(cfg)
	startAll(cfg)

	go watchFiles(cfg)

	helpers.KeepAlive()

	stopAll(cfg)
}

func startAll(cfg processesConfig, only ...string) {

	var wg sync.WaitGroup
	var message []string

	for _, process := range cfg.Processes {

		if process.Enabled && (len(only) == 0 || helpers.SliceHasString(process.Path, only)) {

			wg.Add(1)
			go func(process string) {

				defer wg.Done()

				binPath := TMP + `/bins/GDB_` + process

				cmd := exec.Command("sh", "-c", `go build -o `+binPath+` -ldflags "-X main.version=$(git rev-parse --verify HEAD) -X main.commits=$(git rev-list --count master)" *.go`)
				cmd.Dir = "./cmd/" + process
				err := cmd.Run() // Blocks while building
				if err != nil {
					if exiterr, ok := err.(*exec.ExitError); ok {
						if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
							if status.ExitStatus() == 2 {
								log.Err("Could not build " + process)
								return
							}
						}
					}
					log.ErrS(err, zap.String("process", process))
					return
				}

				cmd = exec.Command("sh", "-c", binPath)
				cmd.Dir = "./cmd/" + process
				err = cmd.Start()
				if err != nil {
					log.ErrS(err, zap.String("process", process))
					return
				}

				err = os.WriteFile(TMP+"/pids/"+process+".pid", []byte(strconv.Itoa(cmd.Process.Pid)), 0644)
				if err != nil {
					log.ErrS(err, zap.String("process", process))
				}

				message = append(message, process)

			}(process.Path)
		}
	}

	wg.Wait()

	if len(message) > 0 {
		log.Info("started: " + strings.Join(message, ", "))
	}
}

func stopAll(cfg processesConfig, only ...string) {

	var wg sync.WaitGroup
	var message []string

	for _, process := range cfg.Processes {

		if process.Enabled && (len(only) == 0 || helpers.SliceHasString(process.Path, only)) {

			wg.Add(1)
			go func(process string) {

				defer wg.Done()

				filename := TMP + "/pids/" + process + ".pid"

				b, err := os.ReadFile(filename)
				if err != nil {
					return
				}

				err = exec.Command("kill", string(b)).Run()
				if err != nil {
					log.ErrS(err, zap.String("process", process))
					return
				}

				err = os.Remove(filename)
				if err != nil {
					log.ErrS(err, zap.String("process", process))
					return
				}

				message = append(message, process)

			}(process.Path)
		}
	}

	wg.Wait()

	if len(message) > 0 {
		log.Info("quited: " + strings.Join(message, ", "))
	}
}

var (
	regxCMDFile      = regexp.MustCompile(`cmd/([a-z]+)/`)
	processesToBuild = map[string]bool{}
	buildTicker      = time.NewTicker(time.Second)
)

func watchFiles(cfg processesConfig) {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.ErrS(err)
		return
	}
	// defer helpers.Close(watcher)

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

						process := regxCMDFile.FindStringSubmatch(event.Name)
						if len(process) == 2 {
							processesToBuild[process[1]] = true
						} else {
							processesToBuild[""] = true
						}

						buildTicker.Reset(time.Second)
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

	if len(dirs) < 10 {
		log.Err("Could not find dirs to watch")
	}

	for _, v := range dirs {
		err = watcher.Add(v)
		if err != nil {
			log.ErrS(err)
		}
	}

	go func() {
		for {

			<-buildTicker.C
			buildTicker.Stop()

			log.InfoS("to build ", processesToBuild)

			if len(processesToBuild) > 0 {

				if _, ok := processesToBuild[""]; ok {
					log.Info("building all")
					stopAll(cfg)
					startAll(cfg)
				} else {
					var x []string
					for k := range processesToBuild {
						x = append(x, k)
					}
					log.InfoS("building ", x)
					stopAll(cfg, x...)
					startAll(cfg, x...)
				}

				processesToBuild = map[string]bool{}

			}
		}
	}()

	<-make(chan bool) // Block
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

//goland:noinspection GoErrorStringFormat
func checkDependencies(cfg processesConfig) error {

	// MySQL
	if cfg.MySQL {
		if !netcat("localhost", "3306") {
			return errors.New("MySQL not running")
		}
	}

	// Memcache
	if cfg.Memcache {
		if !netcat("localhost", "11211") {
			return errors.New("memcache not running")
		}
	}

	// Rabbit
	if cfg.Rabbit {
		if !netcat("localhost", "15672") {
			return errors.New("rabbit not running")
		}
	}

	// LNAV
	if _, err := exec.LookPath("lnav"); err != nil {
		return errors.New("lnav not installed not running")
	}

	return nil
}

func netcat(host, port string) bool {

	cmd := exec.Command("nc", "-vz", host, port)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}

	return strings.Contains(string(out), "open") || // Linux
		strings.Contains(string(out), "succeeded") // OSX
}
