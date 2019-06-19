package queue

import (
	"errors"
	"net/http"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/websockets"
	"github.com/gocolly/colly"
	influx "github.com/influxdata/influxdb1-client"
	"github.com/mitchellh/mapstructure"
	"github.com/powerslacker/ratelimit"
	"github.com/streadway/amqp"
)

var (
	regexIntsOnly = regexp.MustCompile("[^0-9]+")
)

type groupMessage struct {
	ID  string   `json:"id"`
	IDs []string `json:"ids"`
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

	// Backwards compatability, can remove when group queue goes down
	if message.ID != "" {
		message.IDs = append(message.IDs, message.ID)
	}

	defer func(message groupMessage) {
		err := helpers.RemoveKeyFromMemCacheViaPubSub(message.IDs...)
		log.Err(err)
	}(message)

	//
	for _, groupID := range message.IDs {

		var err error
		var group mongo.Group

		group, err = mongo.GetGroup(groupID)
		if err == mongo.ErrNoDocuments || (err == nil && (group.ID64 == "" || group.Type == "")) {

			group.SetID(groupID)
			group.Type, err = getGroupType(groupID)
			if err != nil || group.Type == "" {
				logError(err, groupID)
				payload.ackRetry(msg)
				return
			}

		} else if err != nil {

			logError(err, groupID)
			payload.ackRetry(msg)
			return
		}

		// Continue here means skip this group altogether
		if config.IsProd() && group.UpdatedAt.Unix() > time.Now().Add(time.Hour * -1).Unix() {
			continue
		}

		var found bool
		if group.Type == mongo.GroupTypeGame {
			found, err = updateGameGroup(groupID, &group)
		} else if group.Type == mongo.GroupTypeGroup {
			found, err = updateRegularGroup(groupID, &group)
		} else {
			logError(errors.New("group with no type: "+group.Type), groupID)
			payload.ack(msg)
			return
		}

		if group.ID64 == "" {
			logInfo("Could not find ID64 for: " + groupID)
			payload.ack(msg)
			return
		}

		if err != nil {
			logError(err, groupID)
			payload.ackRetry(msg)
			return
		}

		if !found {
			logWarning("Group counts not found", groupID)
			payload.ackRetry(msg)
			return
		}

		if group.Summary == "No information given." {
			group.Summary = ""
		}

		//
		err = saveGroupToMongo(group)
		if err != nil {
			logError(err, groupID)
			payload.ackRetry(msg)
			return
		}

		//
		err = saveGroupToInflux(group)
		if err != nil {
			logError(err, groupID)
			payload.ackRetry(msg)
			return
		}

		//
		err = sendGroupWebsocket(group)
		if err != nil {
			logError(err, groupID)
		}
	}

	//
	payload.ack(msg)
}

var (
	gameGroupURL = regexp.MustCompile(`steamcommunity\.com/app/([0-9]+)`)
)

func updateGameGroup(id string, group *mongo.Group) (foundNumbers bool, err error) {

	c := colly.NewCollector()

	// ID64
	c.OnHTML("a[href^=\"steam:\"]", func(e *colly.HTMLElement) {
		e.Text = regexIntsOnly.ReplaceAllString(e.Text, "")
		group.ID64 = path.Base(e.Attr("href"))
	})

	// URL
	c.OnHTML("#rightActionBlock .actionItemIcon a", func(e *colly.HTMLElement) {
		matches := gameGroupURL.FindStringSubmatch(e.Attr("href"))
		if len(matches) > 1 {
			group.URL = matches[1]
		}
	})

	// Name
	c.OnHTML("#mainContents > h1", func(e *colly.HTMLElement) {
		group.Name = strings.TrimSpace(e.Text)
	})

	// Headline
	c.OnHTML("#profileBlock > h1", func(e *colly.HTMLElement) {
		group.Headline = strings.TrimSpace(e.Text)
	})

	// Summary
	c.OnHTML("#summaryText", func(e *colly.HTMLElement) {
		var err error
		group.Summary, err = e.DOM.Html()
		log.Err(err)

		if group.Summary == "No information given." {
			group.Summary = ""
		}
	})

	// Icon
	if group.Icon == "" && group.URL != "" {
		i, err := strconv.Atoi(group.URL)
		if err == nil && i > 0 {
			app, err := sql.GetApp(i, []string{"id", "icon"})
			if err != nil {
				log.Err(err)
			}
			group.Icon = app.Icon
		}
	}

	// Members / Members In Chat
	c.OnHTML("#profileBlock .linkStandard", func(e *colly.HTMLElement) {
		if strings.Contains(strings.ToLower(e.Text), "chat") {
			e.Text = regexIntsOnly.ReplaceAllString(e.Text, "")
			group.MembersInChat, err = strconv.Atoi(e.Text)
		} else {
			e.Text = regexIntsOnly.ReplaceAllString(e.Text, "")
			group.Members, err = strconv.Atoi(e.Text)
			foundNumbers = true
		}
	})

	// Members In Game
	c.OnHTML("#profileBlock .membersInGame", func(e *colly.HTMLElement) {
		e.Text = regexIntsOnly.ReplaceAllString(e.Text, "")
		group.MembersInGame, err = strconv.Atoi(e.Text)
	})

	// Members Online
	c.OnHTML("#profileBlock .membersOnline", func(e *colly.HTMLElement) {
		e.Text = regexIntsOnly.ReplaceAllString(e.Text, "")
		group.MembersOnline, err = strconv.Atoi(e.Text)
	})

	// Error
	c.OnHTML("#message h3", func(e *colly.HTMLElement) {
		group.Error = e.Text
		foundNumbers = true
	})

	//
	c.OnError(func(r *colly.Response, err error) {
		log.Err(err)
	})

	return foundNumbers, c.Visit("https://steamcommunity.com/gid/" + id)
}

var (
	regularGroupID64Regex = regexp.MustCompile(`commentthread_Clan_([0-9]{18})_`)
)

func updateRegularGroup(id string, group *mongo.Group) (foundMembers bool, err error) {

	c := colly.NewCollector()

	// ID64
	c.OnHTML("[id^=commentthread_Clan_]", func(e *colly.HTMLElement) {
		matches := regularGroupID64Regex.FindStringSubmatch(e.Attr("id"))
		if len(matches) > 1 {
			group.ID64 = matches[1]
		}
	})

	// Name
	c.OnHTML("div.grouppage_header_name", func(e *colly.HTMLElement) {
		group.Name = strings.TrimPrefix(e.DOM.Children().First().Text(), "/ ")
	})

	// Abbreviation
	c.OnHTML("div.grouppage_header_name span.grouppage_header_abbrev", func(e *colly.HTMLElement) {
		group.Abbr = strings.TrimPrefix(e.Text, "/ ")
	})

	// URL
	c.OnHTML("form#join_group_form", func(e *colly.HTMLElement) {
		group.URL = path.Base(e.Attr("action"))
	})

	// Headline
	c.OnHTML("div.group_content.group_summary h1", func(e *colly.HTMLElement) {
		group.Headline = strings.TrimSpace(e.Text)
	})

	// Summary
	c.OnHTML("div.formatted_group_summary", func(e *colly.HTMLElement) {
		summary, err := e.DOM.Html()
		log.Err(err)
		if err == nil {
			group.Summary = strings.TrimSpace(summary)
		}
	})

	// Icon
	c.OnHTML("div.grouppage_logo img", func(e *colly.HTMLElement) {
		group.Icon = strings.TrimPrefix(e.Attr("src"), mongo.AvatarBase)
	})

	// Members
	c.OnHTML("div.membercount.members .count", func(e *colly.HTMLElement) {
		group.Members, err = strconv.Atoi(regexIntsOnly.ReplaceAllString(e.Text, ""))
		foundMembers = true
	})

	// Members In Game
	c.OnHTML("div.membercount.ingame .count", func(e *colly.HTMLElement) {
		group.MembersInGame, err = strconv.Atoi(regexIntsOnly.ReplaceAllString(e.Text, ""))
	})

	// Members Online
	c.OnHTML("div.membercount.online .count", func(e *colly.HTMLElement) {
		group.MembersOnline, err = strconv.Atoi(regexIntsOnly.ReplaceAllString(e.Text, ""))
	})

	// Members In Chat
	c.OnHTML("div.joinchat_membercount .count", func(e *colly.HTMLElement) {
		group.MembersInChat, err = strconv.Atoi(regexIntsOnly.ReplaceAllString(e.Text, ""))
	})

	// Error
	c.OnHTML("#message h3", func(e *colly.HTMLElement) {
		group.Error = e.Text
		foundMembers = true
	})

	//
	c.OnError(func(r *colly.Response, err error) {
		log.Err(err)
	})

	return foundMembers, c.Visit("https://steamcommunity.com/gid/" + id)
}

func saveGroupToMongo(group mongo.Group) (err error) {

	_, err = mongo.ReplaceDocument(mongo.CollectionGroups, mongo.M{"_id": group.ID64}, group)
	return err
}

func saveGroupToInflux(group mongo.Group) (err error) {

	fields := map[string]interface{}{
		"members_count":   group.Members,
		"members_in_chat": group.MembersInChat,
		"members_in_game": group.MembersInGame,
		"members_online":  group.MembersOnline,
	}

	_, err = helpers.InfluxWrite(helpers.InfluxRetentionPolicyAllTime, influx.Point{
		Measurement: string(helpers.InfluxMeasurementGroups),
		Tags: map[string]string{
			"group_id":   group.ID64,
			"group_type": group.Type,
		},
		Fields:    fields,
		Time:      time.Now(),
		Precision: "h",
	})

	return err
}

func sendGroupWebsocket(group mongo.Group) (err error) {

	// Send websocket
	wsPayload := websockets.PubSubIDStringPayload{} // String as int64 too large for js
	wsPayload.ID = group.ID64
	wsPayload.Pages = []websockets.WebsocketPage{websockets.PageGroup}

	_, err = helpers.Publish(helpers.PubSubTopicWebsockets, wsPayload)
	return err
}

func getGroupType(id string) (string, error) {

	resp, err := http.Get("https://steamcommunity.com/gid/" + id)
	if err != nil {
		return "", err
	}

	defer func() {
		err = resp.Body.Close()
		log.Err(err)
	}()

	u := resp.Request.URL.String()

	if strings.Contains(u, "/games/") {
		return "game", err
	} else if strings.Contains(u, "/groups/") {
		return "group", err
	}

	return "", err
}

var groupXMLRateLimit = ratelimit.New(1, ratelimit.WithCustomDuration(1, time.Minute), ratelimit.WithoutSlack)

//noinspection GoUnusedFunction
func updateGroupFromXML(id string, group *mongo.Group) (err error) {

	groupXMLRateLimit.Take()

	resp, b, err := helpers.GetSteam().GetGroupByID(id)
	err = helpers.HandleSteamStoreErr(err, b, nil)
	if err != nil {
		return err
	}

	group.SetID(id)
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
