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
	ID string `json:"id"`
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
		logInfo("Consuming group: " + message.ID + ", attempt " + strconv.Itoa(payload.Attempt))
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
	if config.IsProd() && group.UpdatedAt.Unix() > time.Now().Add(time.Hour * 1 * -1).Unix() {
		logInfo("Skipping group, updated in last 12 hours")
		payload.ack(msg)
		return
	}

	err = updateGroup(message, &group)
	if err != nil {
		if err.Error() == "expected element type <memberList> but have <html>" {
			logInfo("Group not found", message.ID)
			payload.ack(msg)
		} else {
			logError(err, message.ID)
			payload.ackRetry(msg)
		}
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

	if len(message.ID) < 18 {

		i, err := strconv.ParseInt(message.ID, 10, 32)
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
			"group_id": group.ID64,
		},
		Fields:    fields,
		Time:      time.Now(),
		Precision: "h",
	})

	return err
}
