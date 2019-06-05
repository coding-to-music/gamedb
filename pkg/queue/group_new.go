package queue

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/mitchellh/mapstructure"
	"github.com/powerslacker/ratelimit"
	"github.com/streadway/amqp"
)

type groupQueueNew struct {
	baseQueue
}

func (q groupQueueNew) processMessages(msgs []amqp.Delivery) {

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
	// group, err := mongo.GetGroup(message.ID)
	// log.Info(err)
	// if err != nil && err != mongo.ErrNoDocuments { // Random error, retry
	// 	logError(err, message.ID)
	// 	payload.ackRetry(msg)
	// 	return
	// }
	//
	// if err == nil && group.ID64 != "" { // This can go through normal queue
	// 	log.Info("Putting group back into first queue")
	// 	err = ProduceGroup([]string{message.ID})
	// 	log.Err()
	// 	payload.ack(msg)
	// 	return
	// }

	group := mongo.Group{}

	//
	err = updateGroupFromXML(message.ID, &group)
	if err != nil {
		if err.Error() == "expected element type <memberList> but have <html>" {
			payload.ack(msg)
			return
		} else {
			logError(err, message.ID)
			payload.ackRetry(msg)
			return
		}
	}

	//
	err = saveGroupToMongo(group)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	//
	err = saveGroupToInflux(group)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	//
	err = sendGroupWebsocket(group)
	if err != nil {
		logError(err, message.ID)
	}

	//
	err = helpers.RemoveKeyFromMemCacheViaPubSub(helpers.MemcacheGroup(group.ID64))
	if err != nil {
		logError(err, message.ID)
	}

	err = helpers.RemoveKeyFromMemCacheViaPubSub(helpers.MemcacheGroup(strconv.Itoa(group.ID)))
	if err != nil {
		logError(err, message.ID)
	}

	//
	payload.ack(msg)
}

var groupXMLRateLimit = ratelimit.New(1, ratelimit.WithCustomDuration(1, time.Minute), ratelimit.WithoutSlack)

func updateGroupFromXML(id string, group *mongo.Group) (err error) {

	groupXMLRateLimit.Take()

	resp, b, err := helpers.GetSteam().GetGroupByID(id)
	err = helpers.HandleSteamStoreErr(err, b, nil)
	if err != nil {
		return err
	}

	if len(id) < 18 {

		i, err := strconv.ParseInt(id, 10, 32)
		if err != nil {
			return err
		}
		group.ID = int(i)
	}

	group.ID64 = resp.ID64
	group.Name = resp.Details.Name
	group.URL = resp.Details.URL
	group.Headline = resp.Details.Headline
	group.Summary = resp.Details.Summary
	group.Members = int(resp.Details.MemberCount)
	group.MembersInChat = int(resp.Details.MembersInChat)
	group.MembersInGame = int(resp.Details.MembersInGame)
	group.MembersOnline = int(resp.Details.MembersOnline)
	group.Type = resp.Type

	// Get working icon
	if helpers.GetResponseCode(resp.Details.AvatarFull) == 200 {

		group.Icon = strings.Replace(resp.Details.AvatarFull, mongo.AvatarBase, "", 1)

	} else if helpers.GetResponseCode(resp.Details.AvatarMedium) == 200 {

		group.Icon = strings.Replace(resp.Details.AvatarMedium, mongo.AvatarBase, "", 1)

	} else if helpers.GetResponseCode(resp.Details.AvatarIcon) == 200 {

		group.Icon = strings.Replace(resp.Details.AvatarIcon, mongo.AvatarBase, "", 1)

	} else {

		group.Icon = ""
	}

	return err
}
