package queue

import (
	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type DLCMessage struct {
	AppID  int   `json:"ap_id"`
	DLCIDs []int `json:"dlc_ids"`
}

func appDLCHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		payload := DLCMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			continue
		}

		currentDLCs, err := mongo.GetDLCForApp(0, 0, bson.D{{"app_id", payload.AppID}}, nil)
		if err != nil {
			log.Err(err)
			sendToRetryQueue(message)
			continue
		}

		var currentIDs = map[int]bool{}
		for _, v := range currentDLCs {
			currentIDs[v.DLCID] = true
		}

		var newIDs = map[int]bool{}
		for _, v := range payload.DLCIDs {
			newIDs[v] = true
		}

		// To Add
		var toAdd []int
		for _, v := range payload.DLCIDs {
			if _, ok := currentIDs[v]; !ok {
				toAdd = append(toAdd, v)
			}
		}

		// To Remove
		var toRem []int
		for k := range currentIDs {
			if _, ok := newIDs[k]; !ok {
				toRem = append(toRem, k)
			}
		}

		//
		apps, err := mongo.GetAppsByID(toAdd, bson.M{})
		if err != nil {
			log.Err(err)
			sendToRetryQueue(message)
			continue
		}

		var rows []mongo.AppDLC
		for _, v := range apps {
			rows = append(rows, mongo.AppDLC{
				AppID:           payload.AppID,
				DLCID:           v.ID,
				Icon:            v.Icon,
				Name:            v.Name,
				ReleaseDateNice: v.ReleaseDate,
				ReleaseDateUnix: v.ReleaseDateUnix,
			})
		}

		err = mongo.UpdateAppDLC(rows)
		if err != nil {
			log.Err(err)
			sendToRetryQueue(message)
			continue
		}

		err = mongo.DeleteAppDLC(payload.AppID, toRem)
		if err != nil {
			log.Err(err)
			sendToRetryQueue(message)
			continue
		}

		//
		message.Ack(false)
	}
}
