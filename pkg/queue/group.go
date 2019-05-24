package queue

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/websockets"
	"github.com/gocolly/colly"
	influx "github.com/influxdata/influxdb1-client"
	"github.com/mitchellh/mapstructure"
	"github.com/powerslacker/ratelimit"
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

	// todo, make helper
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
	if config.IsProd() && group.UpdatedAt.Unix() > time.Now().Add(time.Hour * -1).Unix() {
		logInfo("Skipping group, updated recently")
		payload.ack(msg)
		return
	}

	if group.ID64 == "" || group.UpdatedAt.Unix() < time.Now().Add(time.Hour * 24 * 28 * -1).Unix() {

		// Go get details for first time
		err = updateGroupFromXML(message, &group)
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

	} else {

		// Just update counts, scrapes to avoice rate limit
		err = updateGroupFromPage(message, &group)
		if err != nil {
			logError(err, message.ID)
			payload.ackRetry(msg)
			return
		}

	}

	err = addGroupToInflux(group)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	err = sendGroupWebsocket(group)
	if err != nil {
		logError(err, message.ID)
	}

	err = clearGroupMemcache(group)
	if err != nil {
		logError(err, message.ID)
		payload.ackRetry(msg)
		return
	}

	//
	payload.ack(msg)
}

var groupRateLimit = ratelimit.New(1, ratelimit.WithCustomDuration(1, time.Minute), ratelimit.WithoutSlack)

func updateGroupFromXML(message groupMessage, group *mongo.Group) (err error) {

	groupRateLimit.Take()

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

	// Save
	_, err = mongo.ReplaceDocument(mongo.CollectionGroups, mongo.M{"_id": group.ID64}, group)
	return err
}

var reg = regexp.MustCompile(`steamcommunity\.com\/(groups|games|gid)\/`)
var regNums = regexp.MustCompile("[^0-9]+")

func updateGroupFromPage(message groupMessage, group *mongo.Group) (err error) {

	c := colly.NewCollector(
		colly.URLFilters(reg),
	)

	// Regular groups - https://steamcommunity.com/groups/indiegala
	c.OnHTML("div.membercount.members .count", func(e *colly.HTMLElement) {
		e.Text = regNums.ReplaceAllString(e.Text, "")
		group.Members, err = strconv.Atoi(e.Text)
	})

	c.OnHTML("div.membercount.ingame .count", func(e *colly.HTMLElement) {
		e.Text = regNums.ReplaceAllString(e.Text, "")
		group.MembersInGame, err = strconv.Atoi(e.Text)
	})

	c.OnHTML("div.membercount.online .count", func(e *colly.HTMLElement) {
		e.Text = regNums.ReplaceAllString(e.Text, "")
		group.MembersOnline, err = strconv.Atoi(e.Text)
	})

	c.OnHTML("div.joinchat_membercount .count", func(e *colly.HTMLElement) {
		e.Text = regNums.ReplaceAllString(e.Text, "")
		group.MembersInChat, err = strconv.Atoi(e.Text)
	})

	// Game groups - https://steamcommunity.com/games/218620
	c.OnHTML("#profileBlock .linkStandard", func(e *colly.HTMLElement) {
		if strings.Contains(strings.ToLower(e.Text), "chat") {
			e.Text = regNums.ReplaceAllString(e.Text, "")
			group.MembersInChat, err = strconv.Atoi(e.Text)
		} else {
			e.Text = regNums.ReplaceAllString(e.Text, "")
			group.Members, err = strconv.Atoi(e.Text)
		}
	})

	c.OnHTML("#profileBlock .membersInGame", func(e *colly.HTMLElement) {
		e.Text = regNums.ReplaceAllString(e.Text, "")
		group.MembersInGame, err = strconv.Atoi(e.Text)
	})

	c.OnHTML("#profileBlock .membersOnline", func(e *colly.HTMLElement) {
		e.Text = regNums.ReplaceAllString(e.Text, "")
		group.MembersOnline, err = strconv.Atoi(e.Text)
	})

	err = c.Visit("https://steamcommunity.com/gid/" + message.ID)
	if err != nil {
		return err
	}

	// Save
	_, err = mongo.ReplaceDocument(mongo.CollectionGroups, mongo.M{"_id": group.ID64}, group)
	return err
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

func sendGroupWebsocket(group mongo.Group) (err error) {

	// Send websocket
	wsPayload := websockets.PubSubIDStringPayload{}
	wsPayload.ID = group.ID64
	wsPayload.Pages = []websockets.WebsocketPage{websockets.PageGroup}

	_, err = helpers.Publish(helpers.PubSubWebsockets, wsPayload)
	return err
}

func clearGroupMemcache(group mongo.Group) (err error) {

	return nil // Consumers dont have memcache yet
	return helpers.GetMemcache().Delete(helpers.MemcacheGroupRow(group.ID64).Key)
}
