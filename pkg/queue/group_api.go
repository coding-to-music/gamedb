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
	"github.com/streadway/amqp"
)

type groupQueueAPI struct {
}

func (q groupQueueAPI) processMessages(msgs []amqp.Delivery) {

	msg := msgs[0]

	var err error

	message := groupMessage{}
	message.OriginalQueue = queueGroupsNew

	err = helpers.Unmarshal(msg.Body, &message)
	if err != nil {
		logError(err, msg.Body)
		message.fail(msg)
		return
	}

	//
	if len(message.Message.IDs) == 0 {
		log.Err(errors.New("no ids"), msg.Body)
		message.fail(msg)
		return
	}

	if len(message.Message.IDs) > 1 {
		for _, v := range message.Message.IDs {
			err = produceGroupNew(v)
			log.Err(err, msg.Body)
		}
		message.ack(msg)
		return
	}

	var id = message.Message.IDs[0]

	//
	if !helpers.IsValidGroupID(id) {
		log.Err(errors.New("invalid group id: "+id), msg.Body)
		message.fail(msg)
		return
	}

	// See if it's been added
	group, err := mongo.GetGroup(id)
	if err == nil {
		log.Info("Putting group back into first queue")
		err = ProduceGroup([]string{id}, message.Force)
		log.Err(err, msg.Body)
		message.ack(msg)
		return
	} else if err != mongo.ErrNoDocuments {
		log.Err(err)
	}

	//
	var wg sync.WaitGroup

	// Read from steamcommunity.com
	wg.Add(1)
	go func() {

		defer wg.Done()

		var err error

		err = updateGroupFromXML(id, &group)
		if err != nil {

			var ok bool

			// expected element type <memberList> but have <html>
			_, ok = err.(xml.UnmarshalError)
			if ok {
				helpers.LogSteamError(err, id)
				message.ack(msg)
				return
			}

			// XML syntax error on line 7
			_, ok = err.(*xml.SyntaxError)
			if ok {
				helpers.LogSteamError(err, id)
				message.ack(msg)
				return
			}

			helpers.LogSteamError(err, id)
			message.ackRetry(msg)
			return
		}
	}()

	// Read from MySQL
	wg.Add(1)
	var app sql.App
	go func() {

		defer wg.Done()

		var err error

		app, err = getAppFromGroup(group)
		if err != nil {
			logError(err, id)
			message.ackRetry(msg)
			return
		}
	}()

	wg.Wait()

	if message.actionTaken {
		return
	}

	// Save to Mongo
	wg.Add(1)
	go func() {

		defer wg.Done()

		err = saveGroupToMongo(group)
		if err != nil {
			logError(err, id)
			message.ackRetry(msg)
			return
		}
	}()

	// Save to MySQL
	wg.Add(1)
	go func() {

		defer wg.Done()

		err = saveAppsGroupID(app, group.ID64)
		if err != nil {
			logError(err, id)
			message.ackRetry(msg)
			return
		}
	}()

	// Save to Influx
	wg.Add(1)
	go func() {

		defer wg.Done()

		err = saveGroupToInflux(group)
		if err != nil {
			logError(err, id)
			message.ackRetry(msg)
			return
		}
	}()

	wg.Wait()

	if message.actionTaken {
		return
	}

	// Send PubSub
	err = helpers.RemoveKeyFromMemCacheViaPubSub(
		helpers.MemcacheGroup(group.ID64).Key,
		helpers.MemcacheGroup(strconv.Itoa(group.ID)).Key,
	)
	if err != nil {
		logError(err, id)
	}

	//
	err = sendGroupWebsocket([]string{id})
	if err != nil {
		logError(err, id)
	}

	//
	message.ack(msg)
}
