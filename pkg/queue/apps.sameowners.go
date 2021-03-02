package queue

import (
	"math"
	"sort"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type AppSameownersMessage struct {
	AppID int `json:"app_id"`
}

func (m AppSameownersMessage) Queue() rabbit.QueueName {
	return QueueAppsSameowners
}

func appSameownersHandler(message *rabbit.Message) {

	payload := AppSameownersMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err.Error(), zap.String("body", string(message.Message.Body)))
		sendToFailQueue(message)
		return
	}

	// Mark app as done
	success := false
	defer func() {

		if !success {
			return
		}

		var filter = bson.D{{"_id", payload.AppID}}
		var update = bson.D{{"related_owners_app_ids_date", time.Now()}}

		_, err = mongo.UpdateOne(mongo.CollectionApps, filter, update, nil)
		if err != nil {
			log.Err("Updating app", zap.Error(err), zap.Int("app", payload.AppID))
		}
	}()

	//
	ownerRows, err := mongo.GetAppOwners(payload.AppID)
	if err != nil {
		log.Err(err.Error(), zap.Int("app", payload.AppID))
		sendToRetryQueue(message)
		return
	}

	if len(ownerRows) == 0 {
		success = true
		message.Ack()
		return
	}

	var ownerPlayerIDs []int64
	for _, v := range ownerRows {
		ownerPlayerIDs = append(ownerPlayerIDs, v.PlayerID)
	}

	const batch1 = 100
	var countMap = map[int]*mongo.AppSameOwners{}

	for _, chunk := range helpers.ChunkInt64s(ownerPlayerIDs, batch1) {

		ownerApps, err := mongo.GetPlayerAppsByPlayers(chunk, bson.M{"_id": -1, "app_id": 1})
		if err != nil {
			log.Err(err.Error(), zap.Int("app", payload.AppID))
			sendToRetryQueue(message)
			return
		}

		for _, ownerApp := range ownerApps {

			if _, ok := countMap[ownerApp.AppID]; ok {
				countMap[ownerApp.AppID].Count++
			} else {
				countMap[ownerApp.AppID] = &mongo.AppSameOwners{
					SameAppID: ownerApp.AppID,
					Count:     1,
				}
			}
		}
	}

	// Set the order by how popular the game is
	var appIDs []int
	for k := range countMap {
		appIDs = append(appIDs, k)
	}

	const batch2 = 1000
	for _, chunk := range helpers.ChunkInts(appIDs, batch2) {

		apps, err := mongo.GetAppsByID(chunk, bson.M{"owners": 1})
		if err != nil {
			log.Err(err.Error(), zap.Int("app", payload.AppID))
			sendToRetryQueue(message)
			return
		}

		for _, app := range apps {
			if countMap[app.ID].Count > 0 && app.Owners > 0 {
				countMap[app.ID].Order = float64(countMap[app.ID].Count) / math.Pow(float64(app.Owners), 0.5) // Higher removes popular apps
			}
		}
	}

	// Get top 100
	var countSlice []mongo.AppSameOwners
	for _, v := range countMap {
		if v.SameAppID != payload.AppID && v.SameAppID > 0 {
			countSlice = append(countSlice, *v)
		}
	}

	sort.Slice(countSlice, func(i, j int) bool {
		return countSlice[i].Order > countSlice[j].Order
	})

	if len(countSlice) > 100 {
		countSlice = countSlice[0:100]
	}

	// Update same owners table
	err = mongo.ReplaceAppSameOwners(payload.AppID, countSlice)
	if err != nil {
		log.Err(err.Error(), zap.Int("app", payload.AppID))
		sendToRetryQueue(message)
		return
	}

	//
	success = true
	message.Ack()
}
