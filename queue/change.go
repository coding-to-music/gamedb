package queue

import (
	"strconv"
	"strings"

	"github.com/gamedb/website/helpers"
	"github.com/gamedb/website/log"
	"github.com/gamedb/website/mongo"
	"github.com/gamedb/website/sql"
	"github.com/gamedb/website/websockets"
	"github.com/mitchellh/mapstructure"
	"github.com/streadway/amqp"
)

type changeMessage struct {
	ID          int                      `json:"id"`
	PICSChanges RabbitMessageChangesPICS `json:"PICSChanges"`
}

type changeQueue struct {
	baseQueue
}

func (q changeQueue) processMessages(msgs []amqp.Delivery) {

	msg := msgs[0]

	var err error
	var payload = baseMessage{
		Message: changeMessage{},
	}

	err = helpers.Unmarshal(msg.Body, &payload)
	if err != nil {
		logError(err)
		payload.ack(msg)
		return
	}

	var message changeMessage
	err = mapstructure.Decode(payload.Message, &message)
	if err != nil {
		logError(err)
		payload.ack(msg)
		return
	}

	if payload.Attempt > 1 {
		logInfo("Consuming change " + strconv.Itoa(message.ID) + ", attempt " + strconv.Itoa(payload.Attempt))
	}

	// Group products by change ID
	changes := map[int]*mongo.Change{}

	for _, v := range message.PICSChanges.AppChanges {
		if _, ok := changes[v.ChangeNumber]; ok {
			changes[v.ChangeNumber].Apps = append(changes[v.ChangeNumber].Apps, v.ID)
		} else {
			changes[v.ChangeNumber] = &mongo.Change{
				CreatedAt: payload.FirstSeen,
				ID:        v.ChangeNumber,
				Apps:      []int{v.ID},
			}
		}
	}

	for _, v := range message.PICSChanges.PackageChanges {
		if _, ok := changes[v.ChangeNumber]; ok {
			changes[v.ChangeNumber].Packages = append(changes[v.ChangeNumber].Packages, v.ID)
		} else {
			changes[v.ChangeNumber] = &mongo.Change{
				CreatedAt: payload.FirstSeen,
				ID:        v.ChangeNumber,
				Packages:  []int{v.ID},
			}
		}
	}

	// Save to Mongo
	err = saveChangesToMongo(changes)
	if err != nil && !strings.Contains(err.Error(), "duplicate key error collection") {
		logError(err)
		payload.ackRetry(msg)
		return
	}

	// Send websocket
	err = sendChangesWebsocket(changes)
	if err != nil {
		logError(err)
		payload.ackRetry(msg)
		return
	}

	payload.ack(msg)
}

type RabbitMessageChangesPICS struct {
	LastChangeNumber    int  `json:"LastChangeNumber"`
	CurrentChangeNumber int  `json:"CurrentChangeNumber"`
	RequiresFullUpdate  bool `json:"RequiresFullUpdate"`
	PackageChanges      map[string]struct {
		ID           int  `json:"ID"`
		ChangeNumber int  `json:"ChangeNumber"`
		NeedsToken   bool `json:"NeedsToken"`
	} `json:"PackageChanges"`
	AppChanges map[string]struct {
		ID           int  `json:"ID"`
		ChangeNumber int  `json:"ChangeNumber"`
		NeedsToken   bool `json:"NeedsToken"`
	} `json:"AppChanges"`
	JobID steamKitJob `json:"JobID"`
}

func saveChangesToMongo(changes map[int]*mongo.Change) (err error) {

	var changesDocuments []mongo.Document
	for _, v := range changes {

		changesDocuments = append(changesDocuments, mongo.Change{
			ID:        v.ID,
			CreatedAt: v.CreatedAt,
			Apps:      v.Apps,
			Packages:  v.Packages,
		})
	}

	_, err = mongo.InsertDocuments(mongo.CollectionChanges, changesDocuments)
	return err
}

func sendChangesWebsocket(changes map[int]*mongo.Change) (err error) {

	var appIDs []int
	var packageIDs []int
	var appMap = map[int]string{}
	var packageMap = map[int]string{}

	for _, v := range changes {
		appIDs = append(appIDs, v.Apps...)
		packageIDs = append(packageIDs, v.Packages...)
	}

	apps, err := sql.GetAppsByID(appIDs, []string{"id", "name"})
	log.Err(err)

	for _, v := range apps {
		appMap[v.ID] = v.GetName()
	}

	packages, err := sql.GetPackages(packageIDs, []string{"id", "name"})
	log.Err(err)

	for _, v := range packages {
		packageMap[v.ID] = v.GetName()
	}

	page, err := websockets.GetPage(websockets.PageChanges)
	if err != nil {
		return err
	}

	if page.HasConnections() {

		// Make websocket
		var ws [][]interface{}
		for _, v := range changes {

			ws = append(ws, v.OutputForJSON(appMap, packageMap))
		}

		page.Send(ws)
	}

	return nil
}
