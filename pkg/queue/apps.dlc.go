package queue

import (
	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type DLCMessage struct {
	AppID  int   `json:"ap_id"`
	DLCIDs []int `json:"dlc_ids"`
}

func appDLCHandler(message *rabbit.Message) {

	payload := DLCMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err.Error(), zap.ByteString("message", message.Message.Body))
		sendToFailQueue(message)
		return
	}

	currentDLCs, err := mongo.GetDLCForApp(0, 0, bson.D{{"app_id", payload.AppID}}, nil)
	if err != nil {
		log.ErrS(err)
		sendToRetryQueue(message)
		return
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
		log.ErrS(err)
		sendToRetryQueue(message)
		return
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

	err = mongo.ReplaceAppDLCs(rows)
	if err != nil {
		log.ErrS(err)
		sendToRetryQueue(message)
		return
	}

	err = mongo.DeleteAppDLC(payload.AppID, toRem)
	if err != nil {
		log.ErrS(err)
		sendToRetryQueue(message)
		return
	}

	//
	message.Ack()
}
