package queue

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/steam-authority/steam-authority/datastore"
	"github.com/steam-authority/steam-authority/helpers"
	"github.com/steam-authority/steam-authority/logger"
	"github.com/steam-authority/steam-authority/mysql"
	"github.com/steam-authority/steam-authority/websockets"
	"github.com/streadway/amqp"
)

func processChange(msg amqp.Delivery) (ack bool, requeue bool) {

	// Get change
	message := new(RabbitMessageChanges)

	err := json.Unmarshal(msg.Body, message)
	if err != nil {
		if strings.Contains(err.Error(), "cannot unmarshal") {
			logger.Info(err.Error() + " - " + string(msg.Body))
		}

		return false, false
	}

	// Group products by change id
	changes := map[int]*datastore.Change{}

	for _, v := range message.AppChanges {
		if _, ok := changes[v.ChangeNumber]; ok {
			changes[v.ChangeNumber].AddApp(v.ID)
		} else {
			changes[v.ChangeNumber] = &datastore.Change{
				CreatedAt: time.Now(),
				ChangeID:  v.ChangeNumber,
				Apps:      []int{v.ID},
			}
		}
	}

	for _, v := range message.PackageChanges {
		if _, ok := changes[v.ChangeNumber]; ok {
			changes[v.ChangeNumber].AddPackage(v.ID)
		} else {
			changes[v.ChangeNumber] = &datastore.Change{
				CreatedAt: time.Now(),
				ChangeID:  v.ChangeNumber,
				Packages:  []int{v.ID},
			}
		}
	}

	// Make into slice
	var changesSlice []*datastore.Change
	for _, v := range changes {
		changesSlice = append(changesSlice, v)
	}

	// Save change to DS
	err = datastore.BulkAddAChanges(changesSlice)
	if err != nil {
		fmt.Println(err.Error())
		return false, true
	}

	// Send websocket
	if websockets.HasConnections() {

		// Get apps slice
		var appsSlice []int
		for _, v := range message.AppChanges {
			appsSlice = append(appsSlice, v.ID)
		}

		var packagesSlice []int
		for _, v := range message.PackageChanges {
			packagesSlice = append(packagesSlice, v.ID)
		}

		// Get apps for websocket
		appsResp, err := mysql.GetApps(appsSlice, []string{"id", "name"})
		if err != nil {
			logger.Error(err)
		}

		// Make map
		appsRespMap := map[int]mysql.App{}
		for _, v := range appsResp {
			appsRespMap[v.ID] = v
		}

		// Get packages for websocket
		packagesResp, err := mysql.GetPackages(packagesSlice, []string{"id", "name"})
		if err != nil {
			logger.Error(err)
		}

		// Make map
		packagesRespMap := map[int]mysql.Package{}
		for _, v := range packagesResp {
			packagesRespMap[v.ID] = v
		}

		// Make websocket
		ws := websockets.Changes{}
		for _, v := range changes {

			change := websockets.Change{}
			change.ID = v.ChangeID
			change.CreatedAtUnix = v.CreatedAt.Unix()
			change.CreatedAtNice = v.CreatedAt.Format(helpers.DateYearTime)

			for _, appID := range v.Apps {
				if _, ok := appsRespMap[appID]; ok {
					change.AddApp(websockets.ChangeItem{ID: appID, Name: appsRespMap[appID].GetName()})
				} else {
					change.AddApp(websockets.ChangeItem{ID: appID, Name: ""})
				}
			}
			for _, packageID := range v.Packages {
				if _, ok := packagesRespMap[packageID]; ok {
					change.AddApp(websockets.ChangeItem{ID: packageID, Name: packagesRespMap[packageID].GetName()})
				} else {
					change.AddApp(websockets.ChangeItem{ID: packageID, Name: ""})
				}
			}
			ws.AddChange(change)
		}

		websockets.Send(websockets.CHANGES, ws)
	}

	return true, false
}

type RabbitMessageChanges struct {
	LastChangeNumber    int  `json:"LastChangeNumber"`
	CurrentChangeNumber int  `json:"CurrentChangeNumber"`
	RequiresFullUpdate  bool `json:"RequiresFullUpdate"`
	PackageChanges map[string]struct {
		ID           int  `json:"ID"`
		ChangeNumber int  `json:"ChangeNumber"`
		NeedsToken   bool `json:"NeedsToken"`
	} `json:"PackageChanges"`
	AppChanges map[string]struct {
		ID           int  `json:"ID"`
		ChangeNumber int  `json:"ChangeNumber"`
		NeedsToken   bool `json:"NeedsToken"`
	} `json:"AppChanges"`
	JobID struct {
		SequentialCount int    `json:"SequentialCount"`
		StartTime       string `json:"StartTime"`
		ProcessID       int    `json:"ProcessID"`
		BoxID           int    `json:"BoxID"`
		Value           int64  `json:"Value"`
	} `json:"JobID"`
}
