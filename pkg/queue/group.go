package queue

import (
	"encoding/json"
	"net/http"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Jleagle/influxql"
	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	influxHelper "github.com/gamedb/gamedb/pkg/helpers/influx"
	"github.com/gamedb/gamedb/pkg/helpers/memcache"
	pubsubHelpers "github.com/gamedb/gamedb/pkg/helpers/pubsub"
	"github.com/gamedb/gamedb/pkg/helpers/steam"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/gamedb/gamedb/pkg/sql"
	"github.com/gamedb/gamedb/pkg/websockets"
	"github.com/gocolly/colly"
	influx "github.com/influxdata/influxdb1-client"
	"github.com/powerslacker/ratelimit"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	groupScrapeRateLimit = ratelimit.New(1, ratelimit.WithCustomDuration(1, time.Second), ratelimit.WithoutSlack)
)

type GroupMessage struct {
	ID        string  `json:"id"`
	UserAgent *string `json:"user_agent"`
}

func groupsHandler(messages []*rabbit.Message) {

	for _, message := range messages {

		payload := GroupMessage{}

		err := helpers.Unmarshal(message.Message.Body, &payload)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			continue
		}

		payload.ID, err = helpers.UpgradeGroupID(payload.ID)
		if err != nil {
			log.Err(err, message.Message.Body)
			sendToFailQueue(message)
			return
		}

		//
		group, err := mongo.GetGroup(payload.ID)
		if err == mongo.ErrNoDocuments {

			group = mongo.Group{
				ID: payload.ID,
			}

		} else if err != nil {

			log.Err(err, payload.ID)
			sendToRetryQueue(message)
			continue
		}

		// Skip if updated recently
		if !group.ShouldUpdate() {
			message.Ack(false)
			continue
		}

		// Get type/URL
		if group.Type == "" || group.URL == "" {

			group.Type, group.URL, err = getGroupType(payload.ID)
			if err == helpers.ErrInvalidGroupID {

				message.Ack(false)
				continue

			} else if err != nil {

				steam.LogSteamError(err, payload.ID)
				sendToRetryQueue(message)
				continue
			}
		}

		// Update group
		var found bool
		if group.Type == helpers.GroupTypeGame {
			found, err = updateGameGroup(payload.ID, &group)
		} else {
			found, err = updateRegularGroup(payload.ID, &group)
		}

		// Skip if we cant find numbers
		if !found {
			log.Info("Group counts not found", payload.ID)
			sendToRetryQueue(message)
			continue
		}

		// Fix group data
		if group.Summary == "No information given." {
			group.Summary = ""
		}

		// Read from Influx
		err = setGroupTrending(&group)
		if err != nil {
			log.Err(err, payload.ID)
			sendToRetryQueue(message)
			continue
		}

		//
		var wg sync.WaitGroup

		// Read from Mongo
		wg.Add(1)
		var app mongo.App
		go func() {

			defer wg.Done()

			var err error

			app, err = getAppFromGroup(group)
			err = helpers.IgnoreErrors(err, sql.ErrRecordNotFound)
			if err != nil {
				log.Err(err, payload.ID)
				sendToRetryQueue(message)
				return
			}
		}()

		//
		wg.Wait()

		if message.ActionTaken {
			continue
		}

		// Save to Mongo
		wg.Add(1)
		go func() {

			defer wg.Done()

			err = updateApp(app, group)
			if err != nil {
				log.Err(err, payload.ID)
				sendToRetryQueue(message)
				return
			}

			err = saveGroup(group)
			if err != nil {
				log.Err(err, payload.ID)
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
				log.Err(err, payload.ID)
				sendToRetryQueue(message)
				return
			}
		}()

		wg.Wait()

		if message.ActionTaken {
			continue
		}

		// Clear memcache
		err = memcache.RemoveKeyFromMemCacheViaPubSub(payload.ID)
		log.Err(err)

		// Send websocket
		err = sendGroupWebsocket(payload.ID)
		log.Err(err)

		//
		message.Ack(false)
	}
}

func updateGameGroup(id string, group *mongo.Group) (foundNumbers bool, err error) {

	group.Abbr = "" // Game groups don't have abbr's

	groupScrapeRateLimit.Take()

	c := colly.NewCollector(
		colly.AllowURLRevisit(),
	)
	c.SetRequestTimeout(time.Second * 15)

	// ID
	c.OnHTML("a[href^=\"steam:\"]", func(e *colly.HTMLElement) {
		e.Text = helpers.RegexNonInts.ReplaceAllString(e.Text, "")
		group.ID = path.Base(e.Attr("href"))
	})

	// URL
	// c.OnHTML("#eventsBlock a", func(e *colly.HTMLElement) {
	// 	if strings.HasSuffix(e.Attr("href"), "/events") {
	// 		var url = strings.TrimSuffix(e.Attr("href"), "/events")
	// 		group.URL = path.Base(url)
	// 	}
	// })

	// Name
	c.OnHTML("#mainContents > h1", func(e *colly.HTMLElement) {
		var trimmed = strings.TrimSpace(e.Text)
		if trimmed != "" {
			group.Name = trimmed
		}
	})

	// App ID
	c.OnHTML("#rightActionBlock a", func(e *colly.HTMLElement) {
		var url = e.Attr("href")
		if strings.HasSuffix(url, "/discussions") {
			url = strings.TrimSuffix(url, "/discussions")
			url = path.Base(url)
			urli, err := strconv.Atoi(url)
			if err == nil {
				group.AppID = urli
			}
		}
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
			app, err := mongo.GetApp(i, bson.M{"_id": 1, "icon": 1})
			if err != nil {
				log.Err(group.URL, err)
			} else {
				group.Icon = app.Icon
			}
		}
	}

	// Members / Members In Chat
	c.OnHTML("#profileBlock .linkStandard", func(e *colly.HTMLElement) {
		if strings.Contains(strings.ToLower(e.Text), "chat") {
			e.Text = helpers.RegexNonInts.ReplaceAllString(e.Text, "")
			group.MembersInChat, err = strconv.Atoi(e.Text)
		} else {
			e.Text = helpers.RegexNonInts.ReplaceAllString(e.Text, "")
			group.Members, err = strconv.Atoi(e.Text)
			foundNumbers = true
		}
	})

	// Members In Game
	c.OnHTML("#profileBlock .membersInGame", func(e *colly.HTMLElement) {
		e.Text = helpers.RegexNonInts.ReplaceAllString(e.Text, "")
		group.MembersInGame, err = strconv.Atoi(e.Text)
	})

	// Members Online
	c.OnHTML("#profileBlock .membersOnline", func(e *colly.HTMLElement) {
		e.Text = helpers.RegexNonInts.ReplaceAllString(e.Text, "")
		group.MembersOnline, err = strconv.Atoi(e.Text)
	})

	// Error
	group.Error = ""

	c.OnHTML("#message h3", func(e *colly.HTMLElement) {
		group.Error = e.Text
		foundNumbers = true
	})

	//
	c.OnError(func(r *colly.Response, err error) {
		steam.LogSteamError(err)
	})

	return foundNumbers, c.Visit("https://steamcommunity.com/gid/" + id)
}

var (
	regularGroupID64Regex = regexp.MustCompile(`commentthread_Clan_([0-9]{18})_`)
)

func updateRegularGroup(id string, group *mongo.Group) (foundMembers bool, err error) {

	groupScrapeRateLimit.Take()

	group.AppID = 0

	c := colly.NewCollector(
		colly.AllowURLRevisit(),
	)
	c.SetRequestTimeout(time.Second * 60)

	// ID
	c.OnHTML("[id^=commentthread_Clan_]", func(e *colly.HTMLElement) {
		matches := regularGroupID64Regex.FindStringSubmatch(e.Attr("id"))
		if len(matches) > 1 {
			group.ID = matches[1]
		}
	})

	// Abbreviation
	c.OnHTML("div.grouppage_header_name span.grouppage_header_abbrev", func(e *colly.HTMLElement) {
		group.Abbr = strings.TrimPrefix(e.Text, "/ ")
	})

	// Name - Must be after `Abbreviation` as we delete it here.
	c.OnHTML("div.grouppage_header_name", func(e *colly.HTMLElement) {
		e.DOM.Children().Remove()
		var trimmed = strings.TrimSpace(strings.TrimPrefix(e.DOM.Text(), "/ "))
		if trimmed != "" {
			group.Name = trimmed
		}
	})

	// URL
	// c.OnHTML("form#join_group_form", func(e *colly.HTMLElement) {
	// 	group.URL = path.Base(e.Attr("action"))
	// })

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
		group.Icon = strings.TrimPrefix(e.Attr("src"), helpers.AvatarBase)
	})

	// Members
	c.OnHTML("div.membercount.members .count", func(e *colly.HTMLElement) {
		group.Members, err = strconv.Atoi(helpers.RegexNonInts.ReplaceAllString(e.Text, ""))
		foundMembers = true
	})

	// Members In Game
	c.OnHTML("div.membercount.ingame .count", func(e *colly.HTMLElement) {
		group.MembersInGame, err = strconv.Atoi(helpers.RegexNonInts.ReplaceAllString(e.Text, ""))
	})

	// Members Online
	c.OnHTML("div.membercount.online .count", func(e *colly.HTMLElement) {
		group.MembersOnline, err = strconv.Atoi(helpers.RegexNonInts.ReplaceAllString(e.Text, ""))
	})

	// Members In Chat
	c.OnHTML("div.joinchat_membercount .count", func(e *colly.HTMLElement) {
		group.MembersInChat, err = strconv.Atoi(helpers.RegexNonInts.ReplaceAllString(e.Text, ""))
	})

	// Error
	group.Error = ""

	c.OnHTML("#message h3", func(e *colly.HTMLElement) {
		group.Error = e.Text
		foundMembers = true
	})

	//
	c.OnError(func(r *colly.Response, err error) {
		steam.LogSteamError(err)
	})

	return foundMembers, c.Visit("https://steamcommunity.com/gid/" + id)
}

func setGroupTrending(group *mongo.Group) (err error) {

	// Trend value - https://stackoverflow.com/questions/41361734/get-difference-since-30-days-ago-in-influxql-influxdb

	subBuilder := influxql.NewBuilder()
	subBuilder.AddSelect("difference(last(members_count))", "")
	subBuilder.SetFrom(influxHelper.InfluxGameDB, influxHelper.InfluxRetentionPolicyAllTime.String(), influxHelper.InfluxMeasurementGroups.String())
	subBuilder.AddWhere("group_id", "=", group.ID)
	subBuilder.AddWhere("time", ">=", "NOW() - 21d")
	subBuilder.AddGroupByTime("1h")

	builder := influxql.NewBuilder()
	builder.AddSelect("cumulative_sum(difference)", "")
	builder.SetFromSubQuery(subBuilder)

	resp, err := influxHelper.InfluxQuery(builder.String())
	if err != nil {
		return err
	}

	var trendTotal int64

	// Get the last value, todo, put into influx helper, like the ones below
	if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {
		values := resp.Results[0].Series[0].Values
		if len(values) > 0 {

			last := values[len(values)-1]

			trendTotal, err = last[1].(json.Number).Int64()
			if err != nil {
				return err
			}
		}
	}

	group.Trending = trendTotal
	return nil
}

func saveGroup(group mongo.Group) (err error) {

	_, err = mongo.ReplaceOne(mongo.CollectionGroups, bson.D{{"_id", group.ID}}, group)
	if err != nil {
		return err
	}

	// This uses a bunch of cpu
	update := bson.D{
		{"group_name", group.Name},
		{"group_icon", group.Icon},
		{"group_members", group.Members},
		{"group_url", group.URL},
	}

	_, err = mongo.UpdateManySet(mongo.CollectionPlayerGroups, bson.D{{"group_id", group.ID}}, update)
	return err
}

func getAppFromGroup(group mongo.Group) (app mongo.App, err error) {

	if group.Type == helpers.GroupTypeGame && group.AppID > 0 {
		app, err = mongo.GetApp(group.AppID, bson.M{"_id": 1, "group_id": 1})
		if err == mongo.ErrNoDocuments {
			err = ProduceSteam(SteamMessage{AppIDs: []int{group.AppID}})
		}
	}

	return app, err
}

func updateApp(app mongo.App, group mongo.Group) (err error) {

	if app.ID == 0 || group.ID == "" || group.Type != helpers.GroupTypeGame {
		return nil
	}

	_, err = mongo.UpdateOne(mongo.CollectionApps, bson.D{{"_id", app.ID}}, bson.D{
		{"group_id", group.ID},
		{"group_followers", group.Members},
	})
	return err
}

func saveGroupToInflux(group mongo.Group) (err error) {

	fields := map[string]interface{}{
		"members_count":   group.Members,
		"members_in_chat": group.MembersInChat,
		"members_in_game": group.MembersInGame,
		"members_online":  group.MembersOnline,
	}

	_, err = influxHelper.InfluxWrite(influxHelper.InfluxRetentionPolicyAllTime, influx.Point{
		Measurement: string(influxHelper.InfluxMeasurementGroups),
		Tags: map[string]string{
			"group_id":   group.ID,
			"group_type": group.Type,
		},
		Fields:    fields,
		Time:      time.Now(),
		Precision: "h",
	})

	return err
}

func sendGroupWebsocket(id string) (err error) {

	wsPayload := websockets.StringPayload{} // String as int64 too large for js
	wsPayload.String = id
	wsPayload.Pages = []websockets.WebsocketPage{websockets.PageGroup}

	_, err = pubsubHelpers.Publish(pubsubHelpers.PubSubTopicWebsockets, wsPayload)
	return err
}

func getGroupType(id string) (groupType string, groupURL string, err error) {

	groupScrapeRateLimit.Take()

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest("GET", "https://steamcommunity.com/gid/"+id, nil)

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}

	defer func() {
		err = resp.Body.Close()
		log.Err(err)
	}()

	if resp.StatusCode != 302 {
		return "", "", helpers.ErrInvalidGroupID
	}

	redirectURL := resp.Header.Get("Location")

	if strings.Contains(redirectURL, "/games/") {
		return helpers.GroupTypeGame, path.Base(redirectURL), nil
	} else if strings.Contains(redirectURL, "/groups/") {
		return helpers.GroupTypeGroup, path.Base(redirectURL), nil
	}

	return "", "", helpers.ErrInvalidGroupID
}
