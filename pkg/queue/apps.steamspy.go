package queue

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/powerslacker/ratelimit"
	"go.mongodb.org/mongo-driver/bson"
)

type AppSteamspyMessage struct {
	ID int `json:"id"`
}

func (m AppSteamspyMessage) Queue() rabbit.QueueName {
	return QueueAppsSteamspy
}

// https://steamspy.com/api.php
var steamspyLimiter = ratelimit.New(4, ratelimit.WithoutSlack)

func appSteamspyHandler(message *rabbit.Message) {

	payload := AppSteamspyMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		log.Err(err, message.Message.Body)
		sendToFailQueue(message)
		return
	}

	// Create request
	query := url.Values{}
	query.Set("request", "appdetails")
	query.Set("appid", strconv.Itoa(payload.ID))

	steamspyLimiter.Take()
	body, statusCode, err := helpers.GetWithTimeout("https://steamspy.com/api.php?"+query.Encode(), 0)
	if err != nil {

		if strings.Contains(err.Error(), "Client.Timeout exceeded while awaiting headers") ||
			strings.Contains(err.Error(), "read: connection reset by peer") ||
			strings.Contains(err.Error(), "connect: cannot assign requested address") {
			log.Info(err, payload.ID)
		} else {
			log.Err(err, payload.ID)
		}

		sendToRetryQueueWithDelay(message, time.Second*10)
		return
	}

	if statusCode != 200 {

		log.Info(errors.New("steamspy is down"), payload.ID)
		sendToRetryQueueWithDelay(message, time.Minute*30)
		return
	}

	if strings.Contains(string(body), "Connection failed") {

		log.Info(errors.New("steamspy is down"), payload.ID, body)
		sendToRetryQueueWithDelay(message, time.Minute*30)
		return
	}

	// Unmarshal JSON
	resp := mongo.SteamSpyAppResponse{}
	err = helpers.Unmarshal(body, &resp)
	if err != nil {

		log.Info(err, payload.ID, body)
		sendToRetryQueueWithDelay(message, time.Minute*30)
		return
	}

	ss := helpers.AppSteamSpy{}
	ss.SSAveragePlaytimeTwoWeeks = resp.Average2Weeks
	ss.SSAveragePlaytimeForever = resp.AverageForever
	ss.SSMedianPlaytimeTwoWeeks = resp.Median2Weeks
	ss.SSMedianPlaytimeForever = resp.MedianForever

	owners := resp.GetOwners()
	if len(owners) == 2 {
		ss.SSOwnersLow = owners[0]
		ss.SSOwnersHigh = owners[1]
	}

	_, err = mongo.UpdateOne(mongo.CollectionApps, bson.D{{"_id", payload.ID}}, bson.D{{"steam_spy", ss}})
	if err != nil {
		log.Err(err, payload.ID)
		sendToRetryQueue(message)
		return
	}

	err = memcache.Delete(memcache.MemcacheApp(payload.ID).Key)
	if err != nil {
		log.Err(err, payload.ID)
		sendToRetryQueue(message)
		return
	}

	message.Ack(false)
}
