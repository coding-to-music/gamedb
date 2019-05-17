package queue

import (
	"sort"
	"strconv"
	"strings"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/websockets"
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
		Message:       changeMessage{},
		OriginalQueue: queueGoChanges,
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

	// Convert map to slice sor soeting
	var changeSlice []*mongo.Change
	for _, v := range changes {
		changeSlice = append(changeSlice, v)
	}

	sort.Slice(changeSlice, func(i, j int) bool {
		return changeSlice[i].ID < changeSlice[j].ID
	})

	// Save to Mongo
	err = saveChangesToMongo(changeSlice)
	if err != nil && !strings.Contains(err.Error(), "duplicate key error collection") {
		logError(err)
		payload.ackRetry(msg)
		return
	}

	// Get apps and packages for all changes in message
	appMap, packageMap, err := getChangesAppsAndPackages(changeSlice)
	if err != nil {
		logError(err)
		payload.ackRetry(msg)
		return
	}

	// Send websocket
	err = sendChangesWebsocket(changeSlice, appMap, packageMap)
	if err != nil {
		logError(err)
		payload.ackRetry(msg)
		return
	}

	// Send to Discord
	err = sendChangeToDiscord(changeSlice, appMap, packageMap)
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

func saveChangesToMongo(changes []*mongo.Change) (err error) {

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

func getChangesAppsAndPackages(changes []*mongo.Change) (appMap map[int]string, packageMap map[int]string, err error) {

	appMap = map[int]string{}
	packageMap = map[int]string{}

	var appIDs []int
	var packageIDs []int

	for _, v := range changes {
		appIDs = append(appIDs, v.Apps...)
		packageIDs = append(packageIDs, v.Packages...)
	}

	// Apps & packages for all changes
	apps, err := sql.GetAppsByID(appIDs, []string{"id", "name"})
	log.Err(err)

	for _, v := range apps {
		appMap[v.ID] = v.GetName()
	}

	packages, err := sql.GetPackages(packageIDs, []string{"id", "name"})
	if err != nil {
		return appMap, packageMap, err
	}

	for _, v := range packages {
		packageMap[v.ID] = v.GetName()
	}

	return appMap, packageMap, err
}

func sendChangesWebsocket(changes []*mongo.Change, appMap map[int]string, packageMap map[int]string) (err error) {

	var ws [][]interface{}
	for _, v := range changes {

		ws = append(ws, v.OutputForJSON(appMap, packageMap))
	}

	if len(ws) > 0 {

		wsPaload := websockets.PubSubChangesPayload{}
		wsPaload.Data = ws
		wsPaload.Pages = []websockets.WebsocketPage{websockets.PageChanges}

		_, err = helpers.Publish(helpers.PubSubWebsockets, wsPaload)
		log.Err(err)
	}

	return nil
}

func sendChangeToDiscord(changes []*mongo.Change, appMap map[int]string, packageMap map[int]string) (err error) {

	if config.IsProd() {

		discord, err := helpers.GetDiscordBot(config.Config.DiscordChangesBotToken.Get(), true)
		if err != nil {
			return err
		}

		for _, change := range changes {

			var apps []string

			for _, v := range change.Apps {
				if val, ok := appMap[v]; ok {
					apps = append(apps, val)
				}
			}

			if len(apps) > 0 {

				var msg = "Change " + strconv.Itoa(change.ID) + ": " + strings.Join(apps, ", ")
				_, err := discord.ChannelMessageSend("574563721045606431", msg)
				log.Err(err)
			}
		}
	}

	return nil
}
