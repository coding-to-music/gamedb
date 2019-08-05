package queue

import (
	"encoding/xml"
	"errors"
	"strconv"
	"sync"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/mitchellh/mapstructure"
	"github.com/streadway/amqp"
)

type groupQueueAPI struct {
	baseQueue
}

func (q groupQueueAPI) processMessages(msgs []amqp.Delivery) {

	msg := msgs[0]

	var err error
	var payload = baseMessage{
		Message:       groupMessage{},
		OriginalQueue: queueGoGroupsNew,
	}

	err = helpers.Unmarshal(msg.Body, &payload)
	if err != nil {
		logError(err)
		payload.ack(msg)
		return
	}

	var message groupMessage
	err = mapstructure.Decode(payload.Message, &message)
	if err != nil {
		logError(err)
		payload.ack(msg)
		return
	}

	//
	if !helpers.IsValidGroupID(message.ID) {
		log.Err(errors.New("invalid group id: " + message.ID))
		payload.ack(msg)
		return
	}

	// See if it's been added
	group, err := mongo.GetGroup(message.ID)
	if err == nil {
		log.Info("Putting group back into first queue")
		err = ProduceGroup([]string{message.ID})
		log.Err()
		payload.ack(msg)
		return
	} else if err != mongo.ErrNoDocuments {
		log.Err(err)
	}

	//
	var wg sync.WaitGroup

	//
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		err = updateGroupFromXML(message.ID, &group)
		if err != nil {

			var ok bool

			// expected element type <memberList> but have <html>
			_, ok = err.(xml.UnmarshalError)
			if ok {
				helpers.LogSteamError(err, message.ID)
				payload.ack(msg)
				return
			}

			// XML syntax error on line 7
			_, ok = err.(*xml.SyntaxError)
			if ok {
				helpers.LogSteamError(err, message.ID)
				payload.ack(msg)
				return
			}

			helpers.LogSteamError(err, message.ID)
			payload.ackRetry(msg)
			return
		}
	}()

	wg.Add(1)
	var app sql.App
	go func() {

		defer wg.Done()

		var err error

		if group.Type == mongo.GroupTypeGame && group.AppID > 0 {
			app, err = sql.GetApp(group.AppID, []string{"id"})
			if err != nil {
				log.Err(err)
				return
			}

			app.GroupID = group.ID64

			// todo, put this into a save block later on...
			// todo, put into other group queue too
			db, err := sql.GetMySQLClient()
			if err != nil {
				log.Err(err)s
				return
			}

			db = db.Model(&app).Update("group_id", group.ID64)
			if db.Error != nil {
				log.Err(db.Error)
				return
			}
		}
	}()

	wg.Wait()

	if payload.actionTaken {
		return
	}

	wg.Add(1)
	go func() {

		defer wg.Done()

		err = saveGroupToMongo(group)
		if err != nil {
			logError(err, message.ID)
			payload.ackRetry(msg)
			return
		}
	}()

	wg.Add(1)
	go func() {

		defer wg.Done()

		err = saveGroupToInflux(group)
		if err != nil {
			logError(err, message.ID)
			payload.ackRetry(msg)
			return
		}
	}()

	wg.Wait()

	if payload.actionTaken {
		return
	}

	//
	err = helpers.RemoveKeyFromMemCacheViaPubSub(
		helpers.MemcacheGroup(group.ID64).Key,
		helpers.MemcacheGroup(strconv.Itoa(group.ID)).Key,
	)
	if err != nil {
		logError(err, message.ID)
	}

	//
	err = sendGroupWebsocket([]string{message.ID})
	if err != nil {
		logError(err, message.ID)
	}

	//
	payload.ack(msg)
}
