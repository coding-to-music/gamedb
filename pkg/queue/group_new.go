package queue

import (
	"encoding/xml"
	"strconv"
	"strings"
	"sync"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	"github.com/gamedb/gamedb/pkg/helpers/steam"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/queue/framework"
	"github.com/gamedb/gamedb/pkg/sql"
)

type GroupNewMessage struct {
	ID string `json:"id"`
}

func newGroupsHandler(messages []*framework.Message) {

	for _, message := range messages {

		payload := GroupNewMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			return
		}

		//
		payload.ID, err = helpers.UpgradeGroupID(payload.ID)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			return
		}

		// See if it's been added
		group, err := mongo.GetGroup(payload.ID)
		if err == nil {
			log.Info("Putting group back into first queue")
			err = ProduceGroup(GroupMessage{ID: payload.ID})
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
			memcache.MemcacheGroup(group.ID).Key,
		)
		if err != nil {
			log.Err(err, payload.ID)
		}

		//
		err = sendGroupWebsocket(payload.ID)
		if err != nil {
			log.Err(err, payload.ID)
		}

		//
		message.Ack()
	}
}

func updateGroupFromXML(id string, group *mongo.Group) (err error) {

	groupXMLRateLimit.Take()

	resp, b, err := steam.GetSteam().GetGroupByID(id)
	err = steam.AllowSteamCodes(err, b, nil)
	if err != nil {
		return err
	}

	group.ID = id
	group.URL = resp.Details.URL
	group.Headline = resp.Details.Headline
	group.Summary = resp.Details.Summary
	group.Members = int(resp.Details.MemberCount)
	group.MembersInChat = int(resp.Details.MembersInChat)
	group.MembersInGame = int(resp.Details.MembersInGame)
	group.MembersOnline = int(resp.Details.MembersOnline)
	group.Type = resp.Type
	if resp.Details.Name != "" {
		group.Name = resp.Details.Name
	}

	// Try to get App ID from URL
	i, err := strconv.Atoi(resp.Details.URL)
	if err == nil && i > 0 {
		group.AppID = i
	}

	// Get working icon
	if helpers.GetResponseCode(resp.Details.AvatarFull) == 200 {

		group.Icon = strings.Replace(resp.Details.AvatarFull, helpers.AvatarBase, "", 1)

	} else if helpers.GetResponseCode(resp.Details.AvatarMedium) == 200 {

		group.Icon = strings.Replace(resp.Details.AvatarMedium, helpers.AvatarBase, "", 1)

	} else if helpers.GetResponseCode(resp.Details.AvatarIcon) == 200 {

		group.Icon = strings.Replace(resp.Details.AvatarIcon, helpers.AvatarBase, "", 1)

	} else {

		group.Icon = ""
	}

	return nil
}
