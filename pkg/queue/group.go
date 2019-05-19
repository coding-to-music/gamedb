package queue

import (
	"strconv"
	"strings"
	"time"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/mongo"
	influx "github.com/influxdata/influxdb1-client"
	"github.com/mitchellh/mapstructure"
	"github.com/streadway/amqp"
)

type groupMessage struct {
	ID int64 `json:"id"`
}

type groupQueue struct {
	baseQueue
}

func (q groupQueue) processMessages(msgs []amqp.Delivery) {

	msg := msgs[0]

	var err error
	var payload = baseMessage{
		Message:       groupMessage{},
		OriginalQueue: queueGoGroups,
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

	if payload.Attempt > 1 {
		logInfo("Consuming group " + strconv.FormatInt(message.ID, 10) + ", attempt " + strconv.Itoa(payload.Attempt))
	}

	// if !helpers.IsValidAppID(message.ID) {
	// 	logError(errors.New("invalid app ID: " + strconv.Itoa(message.ID)))
	// 	payload.ack(msg)
	// 	return
	// }

	// Load current group
	group, err := mongo.GetGroup(message.ID)
	err = helpers.IgnoreErrors(err, mongo.ErrNoDocuments)
	if err != nil {
		logError(err, message.ID)
		payload.ack(msg)
		return
	}

	// Skip if updated in last day, unless its from PICS
	if config.IsProd() && group.UpdatedAt.Unix() > time.Now().Add(time.Hour * 12 * -1).Unix() {
		logInfo("Skipping group, updated in last 12 hours")
		payload.ack(msg)
		return
	}

	err = updateGroup(message, &group)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	err = addGroupToInflux(group)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	//
	payload.ack(msg)
}

func updateGroup(message groupMessage, group *mongo.Group) (err error) {

	resp, b, err := helpers.GetSteam().GetGroupByID(message.ID)
	err = helpers.HandleSteamStoreErr(err, b, nil)
	if err != nil {
		return err
	}

	group.ID64 = int64(resp.GroupID64)
	group.Name = resp.GroupDetails.GroupName
	group.URL = resp.GroupDetails.GroupURL
	group.Headline = resp.GroupDetails.Headline
	group.Summary = resp.GroupDetails.Summary
	group.Icon = strings.Replace(resp.GroupDetails.AvatarFull, mongo.AvatarBase, "", 1)
	group.Members = int(resp.GroupDetails.MemberCount)
	group.MembersInChat = int(resp.GroupDetails.MembersInChat)
	group.MembersInGame = int(resp.GroupDetails.MembersInGame)
	group.MembersOnline = int(resp.GroupDetails.MembersOnline)

	return group.Save()
}

func addGroupToInflux(group mongo.Group) (err error) {

	fields := map[string]interface{}{
		"members_count":   group.Members,
		"members_in_chat": group.MembersInChat,
		"members_in_game": group.MembersInGame,
		"members_online":  group.MembersOnline,
	}

	_, err = helpers.InfluxWrite(helpers.InfluxRetentionPolicyAllTime, influx.Point{
		Measurement: string(helpers.InfluxMeasurementGroups),
		Tags: map[string]string{
			"group_id": strconv.FormatInt(group.ID64, 10),
		},
		Fields:    fields,
		Time:      time.Now(),
		Precision: "m",
	})

	return err
}
