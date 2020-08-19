package queue

import (
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
var steamspyLimiter = ratelimit.New(4, ratelimit.WithoutSlack, ratelimit.WithCustomDuration(1, time.Minute))

func appSteamspyHandler(message *rabbit.Message) {

	// Disable for now
	message.Ack()
	return

	payload := AppSteamspyMessage{}

	err := helpers.Unmarshal(message.Message.Body, &payload)
	if err != nil {
		zap.L().Error(err.Error(), zap.ByteString("message", message.Message.Body))
		sendToFailQueue(message)
		return
	}

	// Create request
	query := url.Values{}
	query.Set("request", "appdetails")
	query.Set("appid", strconv.Itoa(payload.AppID))

	u := "https://steamspy.com/api.php?" + query.Encode()

	steamspyLimiter.Take()
	body, statusCode, err := helpers.GetWithTimeout(u, 0)
	if err != nil {

		if strings.Contains(err.Error(), "Client.Timeout exceeded while awaiting headers") ||
			strings.Contains(err.Error(), "read: connection reset by peer") ||
			strings.Contains(err.Error(), "connect: cannot assign requested address") {
			zap.S().Info(err, payload.AppID)
		} else {
			zap.S().Error(err, payload.AppID, u)
		}

		sendToRetryQueueWithDelay(message, time.Second*10)
		return
	}

	if statusCode != 200 {

		zap.S().Info("steamspy is down", payload.AppID)
		sendToRetryQueueWithDelay(message, time.Minute*30)
		return
	}

	if strings.Contains(string(body), "Connection failed") {

		zap.L().Info("steamspy is down", zap.Int("app", payload.AppID), zap.ByteString("bytes", body))
		sendToRetryQueueWithDelay(message, time.Minute*30)
		return
	}

	// Unmarshal JSON
	resp := mongo.SteamSpyAppResponse{}
	err = helpers.Unmarshal(body, &resp)
	if err != nil {

		zap.S().Info(err, payload.AppID, helpers.TruncateString(string(body), 200, "..."))
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
		zap.S().Error(err, payload.AppID, u)
		sendToRetryQueue(message)
		return
	}

	// Clear cache
	err = memcache.Delete(memcache.MemcacheApp(payload.AppID).Key)
	if err != nil {
		zap.S().Error(err, payload.AppID, u)
		sendToRetryQueue(message)
		return
	}

	// Update in Elastic
	err = ProduceAppSearch(nil, payload.AppID)
	if err != nil {
		zap.S().Error(err, payload.AppID, u)
		sendToRetryQueue(message)
		return
	}

	//
	message.Ack()
}
