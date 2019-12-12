package consumers

import (
	"encoding/xml"
	"strconv"
	"sync"

	"github.com/gamedb/gamedb/pkg/consumers/framework"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/helpers/steam"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
)

type GroupSingleMessage struct {
	ID string `json:"id"`
}

func newGroupsHandler(messages []*framework.Message) {

	for _, message := range messages {

		payload := GroupSingleMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			return
		}

		//
		if !helpers.IsValidGroupID(payload.ID) {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			return
		}

		// See if it's been added
		group, err := mongo.GetGroup(payload.ID)
		if err == nil {
			log.Info("Putting group back into first queue")
			err = ProduceGroup(GroupMessage{IDs: []string{payload.ID}})
			if err != nil {
				log.Err(err, message.Message.Body)
				sendToRetryQueue(message)
			} else {
				message.Ack()
			}
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

			err = updateGroupFromXML(payload.ID, &group)
			if err != nil {

				var ok bool

				// expected element type <memberList> but have <html>
				_, ok = err.(xml.UnmarshalError)
				if ok {
					steam.LogSteamError(err, message.Message.Body)
					message.Ack()
					return
				}

				// XML syntax error on line 7
				_, ok = err.(*xml.SyntaxError)
				if ok {
					steam.LogSteamError(err, message.Message.Body)
					message.Ack()
					return
				}

				steam.LogSteamError(err, message.Message.Body)
				sendToRetryQueue(message)
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
				log.Err(err, message.Message.Body)
				sendToRetryQueue(message)
				return
			}
		}()

		wg.Wait()

		if message.ActionTaken {
			continue
		}

		// Save to MySQL
		wg.Add(1)
		go func() {

			defer wg.Done()

			err = saveAppsGroupID(app, group)
			if err != nil {
				log.Err(err, message.Message.Body)
				sendToRetryQueue(message)
				return
			}
		}()

		// Save to Mongo
		wg.Add(1)
		go func() {

			defer wg.Done()

			err = saveGroup(group)
			if err != nil {
				log.Err(err, message.Message.Body)
				sendToRetryQueue(message)
				return
			}
		}()

		// Save to Influx
		wg.Add(1)
		go func() {

			defer wg.Done()

			err = saveGroupToInflux(group)
			if err != nil {
				log.Err(err, message.Message.Body)
				sendToRetryQueue(message)
				return
			}
		}()

		wg.Wait()

		if message.ActionTaken {
			continue
		}

		// Send PubSub
		err = memcache.RemoveKeyFromMemCacheViaPubSub(
			memcache.MemcacheGroup(group.ID64).Key,
			memcache.MemcacheGroup(strconv.Itoa(group.ID)).Key,
		)
		if err != nil {
			log.Err(err, payload.ID)
		}

		//
		err = sendGroupWebsocket([]string{payload.ID})
		if err != nil {
			log.Err(err, payload.ID)
		}

		//
		message.Ack()
	}
}
