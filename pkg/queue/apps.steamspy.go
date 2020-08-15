package queue

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Jleagle/rabbit-go"
	"github.com/gamedb/gamedb/pkg/helpers"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/mongo"
	"github.com/powerslacker/ratelimit"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type AppSteamspyMessage struct {
	AppID int `json:"id"`
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
		zap.S().Error(err, message.Message.Body)
		sendToFailQueue(message)
		return
	}

	// Create request
	query := url.Values{}
	query.Set("request", "appdetails")
	query.Set("appid", strconv.Itoa(payload.AppID))

	steamspyLimiter.Take()
	body, statusCode, err := helpers.GetWithTimeout("https://steamspy.com/api.php?"+query.Encode(), 0)
	if err != nil {

		if strings.Contains(err.Error(), "Client.Timeout exceeded while awaiting headers") ||
			strings.Contains(err.Error(), "read: connection reset by peer") ||
			strings.Contains(err.Error(), "connect: cannot assign requested address") {
			zap.S().Info(err, payload.AppID)
		} else {
			zap.S().Error(err, payload.AppID)
		}

		sendToRetryQueueWithDelay(message, time.Second*10)
		return
	}

	if statusCode != 200 {

		zap.S().Info(errors.New("steamspy is down"), payload.AppID)
		sendToRetryQueueWithDelay(message, time.Minute*30)
		return
	}

	if strings.Contains(string(body), "Connection failed") {

		zap.S().Info(errors.New("steamspy is down"), payload.AppID, body)
		sendToRetryQueueWithDelay(message, time.Minute*30)
		return
	}

	// Unmarshal JSON
	resp := mongo.SteamSpyAppResponse{}
	err = helpers.Unmarshal(body, &resp)
	if err != nil {

		zap.S().Info(err, payload.AppID, body)
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

	// Update app row
	filter := bson.D{{"_id", payload.AppID}}
	update := bson.D{{"steam_spy", ss}}

	_, err = mongo.UpdateOne(mongo.CollectionApps, filter, update)
	if err != nil {
		zap.S().Error(err, payload.AppID)
		sendToRetryQueue(message)
		return
	}

	// Clear cache
	err = memcache.Delete(memcache.MemcacheApp(payload.AppID).Key)
	if err != nil {
		zap.S().Error(err, payload.AppID)
		sendToRetryQueue(message)
		return
	}

	// Update in Elastic
	err = ProduceAppSearch(nil, payload.AppID)
	if err != nil {
		zap.S().Error(err, payload.AppID)
		sendToRetryQueue(message)
		return
	}

	//
	message.Ack()
}
