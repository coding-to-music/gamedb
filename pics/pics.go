package pics

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/queue"
)

const (
	changesLimit = 500
	checkSeconds = 10
	bigChangeID  = 4067165 // Fallback when there are no changes in DB
)

var (
	latestChangeSaved int
	enablePics        bool
)

// Run triggers the PICS updater to run forever
func Run() {

	for {
		jsChange, err := getLatestChanges()
		if err != nil {

			if strings.HasSuffix(err.Error(), "connect: connection refused") {
				time.Sleep(time.Second)
				continue
			}
		}

		for k, v := range jsChange.Apps {
			appID, _ := strconv.Atoi(k)
			bytes, _ := json.Marshal(queue.AppMessage{
				AppID:    appID,
				ChangeID: v,
			})

			queue.Produce(queue.AppQueue, bytes)
		}

		for k, v := range jsChange.Packages {
			packageID, _ := strconv.Atoi(k)
			bytes, _ := json.Marshal(queue.PackageMessage{
				PackageID: packageID,
				ChangeID:  v,
			})

			queue.Produce(queue.PackageQueue, bytes)
		}

		// Make a list of changes to add
		changes := make(map[int]*datastore.Change, 0)

		// todo, check if the change already exists, if so, no need to do this.
		for k, v := range jsChange.Apps {
			_, ok := changes[v]
			if !ok {
				changes[v] = &datastore.Change{
					ChangeID:  v,
					CreatedAt: time.Now(),
				}
			}

			intx, _ := strconv.Atoi(k)
			changes[v].Apps = append(changes[v].Apps, intx)
		}

		for k, v := range jsChange.Packages {
			_, ok := changes[v]
			if !ok {
				changes[v] = &datastore.Change{
					ChangeID:  v,
					CreatedAt: time.Now(),
				}
			}

			intx, _ := strconv.Atoi(k)
			changes[v].Packages = append(changes[v].Packages, intx)
		}

		// Add changes to rabbit
		for _, v := range changes {
			bytes, _ := json.Marshal(*v)
			queue.Produce(queue.ChangeQueue, bytes)
		}

		// Sleep
		time.Sleep(time.Second * checkSeconds)
	}
}

func getLatestChanges() (jsChange JsChange, err error) {

	// Get the last change
	if latestChangeSaved == 0 {

		changes, err := datastore.GetLatestChanges(1)
		if err != nil {
			logger.Error(err)
		}

		if len(changes) > 0 {
			latestChangeSaved = changes[0].ChangeID
		} else {
			latestChangeSaved = bigChangeID
		}
	}

	// Grab the JSON from node
	url := "http://localhost:8086/changes/" + strconv.Itoa(latestChangeSaved)
	//logger.Info("PICS: " + url)
	response, err := http.Get(url)
	if err != nil {
		return jsChange, err
	}
	defer response.Body.Close()

	// Convert to bytes
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return jsChange, err
	}

	// Unmarshal JSON
	if err := json.Unmarshal(contents, &jsChange); err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			logger.Info(err.Error() + " - " + string(contents))
		}
		return jsChange, err
	}

	latestChangeSaved = jsChange.LatestChangeNumber

	return jsChange, nil
}

type JsChange struct {
	Success            int8           `json:"success"`
	LatestChangeNumber int            `json:"current_changenumber"`
	Apps               map[string]int `json:"apps"`
	Packages           map[string]int `json:"packages"`
}
