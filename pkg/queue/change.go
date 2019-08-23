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
	AppIDs     map[int]int `json:"app_ids"`
	PackageIDs map[int]int `json:"package_ids"`
}

type changeQueue struct {
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

	// Group products by change ID
	changes := map[int]*mongo.Change{}

	for changeNumber, appID := range message.AppIDs {
		if _, ok := changes[changeNumber]; ok {
			changes[changeNumber].Apps = append(changes[changeNumber].Apps, appID)
		} else {
			changes[changeNumber] = &mongo.Change{
				CreatedAt: payload.FirstSeen,
				ID:        changeNumber,
				Apps:      []int{appID},
			}
		}
	}

	for changeNumber, packageID := range message.PackageIDs {
		if _, ok := changes[changeNumber]; ok {
			changes[changeNumber].Packages = append(changes[changeNumber].Packages, packageID)
		} else {
			changes[changeNumber] = &mongo.Change{
				CreatedAt: payload.FirstSeen,
				ID:        changeNumber,
				Packages:  []int{packageID},
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

		_, err = helpers.Publish(helpers.PubSubTopicWebsockets, wsPaload)
		log.Err(err)
	}

	return nil
}

// todo, add packages to return
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
